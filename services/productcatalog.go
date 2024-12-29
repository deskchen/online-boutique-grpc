package services

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/appnetorg/OnlineBoutique/protos/productcatalog"
)

// NewProductCatalogService returns a new server for the ProductCatalogService
func NewProductCatalogService(port int) *ProductCatalogService {
	return &ProductCatalogService{
		name:     "product-catalog-service",
		port:     port,
		products: make(map[string]*productcatalog.Product), // In-memory storage for simplicity
	}
}

// ProductCatalogService implements the ProductCatalogService
type ProductCatalogService struct {
	name string
	port int
	productcatalog.ProductCatalogServiceServer
	products map[string]*productcatalog.Product // Mock in-memory storage for products
}

// Run starts the server
func (s *ProductCatalogService) Run() error {
	srv := grpc.NewServer()
	productcatalog.RegisterProductCatalogServiceServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("ProductCatalogService running at port: %d", s.port)
	return srv.Serve(lis)
}

// ListProducts lists all available products
func (s *ProductCatalogService) ListProducts(ctx context.Context, req *productcatalog.Empty) (*productcatalog.ListProductsResponse, error) {
	log.Printf("ListProducts request received")

	var productList []*productcatalog.Product
	for _, product := range s.products {
		productList = append(productList, product)
	}

	return &productcatalog.ListProductsResponse{
		Products: productList,
	}, nil
}

// GetProduct retrieves a product by its ID
func (s *ProductCatalogService) GetProduct(ctx context.Context, req *productcatalog.GetProductRequest) (*productcatalog.Product, error) {
	log.Printf("GetProduct request for id = %v", req.GetId())

	product, exists := s.products[req.GetId()]
	if !exists {
		return nil, fmt.Errorf("product with id %v not found", req.GetId())
	}

	return product, nil
}

// SearchProducts searches for products matching a query
func (s *ProductCatalogService) SearchProducts(ctx context.Context, req *productcatalog.SearchProductsRequest) (*productcatalog.SearchProductsResponse, error) {
	log.Printf("SearchProducts request with query = %v", req.GetQuery())

	var results []*productcatalog.Product
	for _, product := range s.products {
		if contains(product.Name, req.GetQuery()) || contains(product.Description, req.GetQuery()) {
			results = append(results, product)
		}
	}

	return &productcatalog.SearchProductsResponse{
		Results: results,
	}, nil
}

// contains checks if a string contains a substring (case-insensitive)
func contains(source, query string) bool {
	return len(source) >= len(query) && (len(query) == 0 ||
		len(source) == len(query) && source == query ||
		len(source) > len(query) && (source[:len(query)] == query || contains(source[1:], query)))
}
