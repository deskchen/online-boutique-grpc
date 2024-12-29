package services

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/appnetorg/OnlineBoutique/protos/ad"
)

// NewAdService returns a new server for the AdService
func NewAdService(port int) *AdService {
	return &AdService{
		name: "ad-service",
		port: port,
	}
}

// AdService implements the AdService
type AdService struct {
	name string
	port int
	ad.AdServiceServer
}

// Run starts the server
func (s *AdService) Run() error {
	opts := []grpc.ServerOption{
		// grpc.UnaryInterceptor(
		// ),
	}

	srv := grpc.NewServer(opts...)
	ad.RegisterAdServiceServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("AdService running at port: %d", s.port)
	return srv.Serve(lis)
}

// GetAds returns a list of ads based on the context keys
func (s *AdService) GetAds(ctx context.Context, req *ad.AdRequest) (*ad.AdResponse, error) {
	log.Printf("GetAds request with context_keys = %v", req.GetContextKeys())

	// Mock data for ads. Replace with actual business logic.
	ads := []*ad.Ad{
		{
			RedirectUrl: "https://example.com/product1",
			Text:        "Buy the best product 1!",
		},
		{
			RedirectUrl: "https://example.com/product2",
			Text:        "Get a discount on product 2!",
		},
	}

	return &ad.AdResponse{
		Ads: ads,
	}, nil
}
