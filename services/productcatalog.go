package services

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/appnetorg/OnlineBoutique/protos/onlineboutique"
	"github.com/golang/protobuf/jsonpb"
)

// ProductCatalogService implements the ProductCatalogService
type ProductCatalogService struct {
	port    int
	catalog pb.ListProductsResponse

	pb.ProductCatalogServiceServer

	mu            sync.RWMutex
	extraLatency  time.Duration
	reloadCatalog bool
}

// NewProductCatalogService creates a new ProductCatalogService
func NewProductCatalogService(port int) *ProductCatalogService {
	svc := &ProductCatalogService{
		port: port,
	}

	// Initialize extra latency from environment variable
	if extra := os.Getenv("EXTRA_LATENCY"); extra != "" {
		if duration, err := time.ParseDuration(extra); err == nil {
			svc.extraLatency = duration
		}
	}

	// Signal handling for reloading
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR1, syscall.SIGUSR2)
	go func() {
		for {
			sig := <-sigs
			if sig == syscall.SIGUSR1 {
				log.Println("Enabling catalog reload")
				svc.mu.Lock()
				svc.reloadCatalog = true
				svc.mu.Unlock()
			} else if sig == syscall.SIGUSR2 {
				log.Println("Disabling catalog reload")
				svc.mu.Lock()
				svc.reloadCatalog = false
				svc.mu.Unlock()
			}
		}
	}()

	// Load initial catalog
	if err := svc.loadCatalog(&svc.catalog); err != nil {
		log.Fatalf("Failed to load catalog: %v", err)
	}

	return svc
}

// loadCatalog loads the product catalog from a file
func (s *ProductCatalogService) loadCatalog(catalog *pb.ListProductsResponse) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	catalogJSON, err := os.ReadFile("data/products.json")
	if err != nil {
		return err
	}
	if err := jsonpb.Unmarshal(bytes.NewReader(catalogJSON), catalog); err != nil {
		return err
	}

	return nil
}

// parseCatalog parses the current catalog state
func (s *ProductCatalogService) parseCatalog() []*pb.Product {
	if s.reloadCatalog || len(s.catalog.Products) == 0 {
		err := s.loadCatalog(&s.catalog)
		if err != nil {
			return []*pb.Product{}
		}
	}

	return s.catalog.Products
}

// Run starts the gRPC server
func (s *ProductCatalogService) Run() error {
	srv := grpc.NewServer()
	pb.RegisterProductCatalogServiceServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("ProductCatalogService running at port: %d", s.port)
	return srv.Serve(lis)
}

// ListProducts lists all available products
func (s *ProductCatalogService) ListProducts(ctx context.Context, req *pb.Empty) (*pb.ListProductsResponse, error) {
	time.Sleep(s.extraLatency)

	return &pb.ListProductsResponse{
		Products: s.parseCatalog(),
	}, nil
}

// GetProduct retrieves a product by its ID
func (s *ProductCatalogService) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
	time.Sleep(s.extraLatency)

	var found *pb.Product
	for i := 0; i < len(s.parseCatalog()); i++ {
		if req.Id == s.parseCatalog()[i].Id {
			found = s.parseCatalog()[i]
		}
	}

	if found == nil {
		return nil, status.Errorf(codes.NotFound, "no product with ID %s", req.Id)
	}

	return found, nil
}

// SearchProducts searches for products matching a query
func (s *ProductCatalogService) SearchProducts(ctx context.Context, req *pb.SearchProductsRequest) (*pb.SearchProductsResponse, error) {
	time.Sleep(s.extraLatency)

	var ps []*pb.Product
	for _, product := range s.parseCatalog() {
		if strings.Contains(strings.ToLower(product.Name), strings.ToLower(req.Query)) ||
			strings.Contains(strings.ToLower(product.Description), strings.ToLower(req.Query)) {
			ps = append(ps, product)
		}
	}

	return &pb.SearchProductsResponse{Results: ps}, nil
}
