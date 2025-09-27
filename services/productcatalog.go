package services

import (
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
	"google.golang.org/protobuf/encoding/protojson"

	pb "github.com/deskchen/online-boutique-grpc/protos/onlineboutique"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
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
			switch sig {
			case syscall.SIGUSR1:
				log.Println("Enabling catalog reload")
				svc.mu.Lock()
				svc.reloadCatalog = true
				svc.mu.Unlock()
			case syscall.SIGUSR2:
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

// loadCatalog loads the product catalog from a file.
func (s *ProductCatalogService) loadCatalog(catalog *pb.ListProductsResponse) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Read the JSON file
	catalogJSON, err := os.ReadFile("data/products.json")
	if err != nil {
		return err
	}

	// Unmarshal the JSON into the Protobuf message
	if err := protojson.Unmarshal(catalogJSON, catalog); err != nil {
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
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(opentracing.GlobalTracer())),
	}
	srv := grpc.NewServer(opts...)
	pb.RegisterProductCatalogServiceServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("ProductCatalogService running at port: %d", s.port)
	return srv.Serve(lis)
}

// ListProducts lists all available products
func (s *ProductCatalogService) ListProducts(ctx context.Context, req *pb.EmptyUser) (*pb.ListProductsResponse, error) {
	log.Println("ListProducts: Received request")

	time.Sleep(s.extraLatency)

	response := &pb.ListProductsResponse{
		Products: s.parseCatalog(),
	}

	log.Printf("ListProducts: Responding with %d products\n", len(response.Products))

	return response, nil
}

// GetProduct retrieves a product by its ID
func (s *ProductCatalogService) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
	log.Printf("GetProduct: Received request for product ID %s\n", req.Id)

	time.Sleep(s.extraLatency)

	var found *pb.Product
	for i := 0; i < len(s.parseCatalog()); i++ {
		if req.Id == s.parseCatalog()[i].Id {
			found = s.parseCatalog()[i]
			break
		}
	}

	if found == nil {
		log.Printf("GetProduct: Product with ID %s not found\n", req.Id)
		return nil, status.Errorf(codes.NotFound, "no product with ID %s", req.Id)
	}

	log.Printf("GetProduct: Found product with ID %s\n", found.Id)
	return found, nil
}

// SearchProducts searches for products matching a query
func (s *ProductCatalogService) SearchProducts(ctx context.Context, req *pb.SearchProductsRequest) (*pb.SearchProductsResponse, error) {
	log.Printf("SearchProducts: Received request with query: %s\n", req.Query)

	time.Sleep(s.extraLatency)

	var ps []*pb.Product
	for _, product := range s.parseCatalog() {
		if strings.Contains(strings.ToLower(product.Name), strings.ToLower(req.Query)) ||
			strings.Contains(strings.ToLower(product.Description), strings.ToLower(req.Query)) {
			ps = append(ps, product)
		}
	}

	log.Printf("SearchProducts: Search completed. Query: %s, Results: %d\n", req.Query, len(ps))

	return &pb.SearchProductsResponse{Results: ps}, nil
}
