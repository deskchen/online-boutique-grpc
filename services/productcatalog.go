package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"

	pb "github.com/appnetorg/OnlineBoutique/protos/onlineboutique"
)

// Product represents a product in the catalog
type Product struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Picture     string   `json:"picture"`
	PriceUSD    string   `json:"priceUsd"` // Adjust type to match your proto if needed
	Categories  []string `json:"categories"`
}

// ProductCatalogService implements the ProductCatalogService
type ProductCatalogService struct {
	port     int
	products []Product // Catalog storage

	pb.ProductCatalogServiceServer

	mu            sync.RWMutex
	extraLatency  time.Duration
	reloadCatalog bool
}

// NewProductCatalogService creates a new ProductCatalogService
func NewProductCatalogService(port int) *ProductCatalogService {
	svc := &ProductCatalogService{
		port:     port,
		products: []Product{},
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
	if err := svc.loadCatalog(); err != nil {
		log.Fatalf("Failed to load catalog: %v", err)
	}

	return svc
}

// loadCatalog loads the product catalog from a file
func (s *ProductCatalogService) loadCatalog() error {
	catalogData, err := os.ReadFile("data/products.json")
	if err != nil {
		return err
	}

	var products []Product
	if err := json.Unmarshal(catalogData, &products); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.products = products
	return nil
}

// parseCatalog parses the current catalog state
func (s *ProductCatalogService) parseCatalog() []Product {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.reloadCatalog {
		if err := s.loadCatalog(); err != nil {
			log.Printf("Failed to reload catalog: %v", err)
		}
	}
	return s.products
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

	var productList []*pb.Product
	for _, p := range s.parseCatalog() {
		productList = append(productList, &pb.Product{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Picture:     p.Picture,
			PriceUsd:    &pb.Money{CurrencyCode: "USD", Units: parsePrice(p.PriceUSD)},
			Categories:  p.Categories,
		})
	}

	return &pb.ListProductsResponse{
		Products: productList,
	}, nil
}

// GetProduct retrieves a product by its ID
func (s *ProductCatalogService) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
	time.Sleep(s.extraLatency)

	for _, p := range s.parseCatalog() {
		if p.ID == req.GetId() {
			return &pb.Product{
				Id:          p.ID,
				Name:        p.Name,
				Description: p.Description,
				Picture:     p.Picture,
				PriceUsd:    &pb.Money{CurrencyCode: "USD", Units: parsePrice(p.PriceUSD)},
				Categories:  p.Categories,
			}, nil
		}
	}

	return nil, fmt.Errorf("product with id %v not found", req.GetId())
}

// SearchProducts searches for products matching a query
func (s *ProductCatalogService) SearchProducts(ctx context.Context, req *pb.SearchProductsRequest) (*pb.SearchProductsResponse, error) {
	time.Sleep(s.extraLatency)

	var results []*pb.Product
	for _, p := range s.parseCatalog() {
		if strings.Contains(strings.ToLower(p.Name), strings.ToLower(req.GetQuery())) ||
			strings.Contains(strings.ToLower(p.Description), strings.ToLower(req.GetQuery())) {
			results = append(results, &pb.Product{
				Id:          p.ID,
				Name:        p.Name,
				Description: p.Description,
				Picture:     p.Picture,
				PriceUsd:    &pb.Money{CurrencyCode: "USD", Units: parsePrice(p.PriceUSD)},
				Categories:  p.Categories,
			})
		}
	}

	return &pb.SearchProductsResponse{
		Results: results,
	}, nil
}

// parsePrice converts a price string to an int64
func parsePrice(price string) int64 {
	parsedPrice, err := strconv.ParseFloat(price, 64)
	if err != nil {
		log.Printf("Failed to parse price: %v", err)
		return 0
	}
	return int64(parsedPrice)
}
