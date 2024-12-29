package services

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"

	"google.golang.org/grpc"

	"github.com/appnetorg/OnlineBoutique/protos/recommendation"
)

// NewRecommendationService returns a new server for the RecommendationService
func NewRecommendationService(port int) *RecommendationService {
	return &RecommendationService{
		name: "recommendation-service",
		port: port,
	}
}

// RecommendationService implements the RecommendationService
type RecommendationService struct {
	name string
	port int
	recommendation.RecommendationServiceServer
}

// Run starts the server
func (s *RecommendationService) Run() error {
	srv := grpc.NewServer()
	recommendation.RegisterRecommendationServiceServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("RecommendationService running at port: %d", s.port)
	return srv.Serve(lis)
}

// ListRecommendations provides a list of recommended product IDs based on user and product history
func (s *RecommendationService) ListRecommendations(ctx context.Context, req *recommendation.ListRecommendationsRequest) (*recommendation.ListRecommendationsResponse, error) {
	log.Printf("ListRecommendations request received for user_id = %v, product_ids = %v", req.GetUserId(), req.GetProductIds())

	// Mock recommendation logic: shuffle input product IDs and return a subset
	productIDs := req.GetProductIds()
	rand.Shuffle(len(productIDs), func(i, j int) { productIDs[i], productIDs[j] = productIDs[j], productIDs[i] })

	recommended := []string{}
	if len(productIDs) > 0 {
		maxRecommendations := 5
		if len(productIDs) < maxRecommendations {
			maxRecommendations = len(productIDs)
		}
		recommended = productIDs[:maxRecommendations]
	}

	return &recommendation.ListRecommendationsResponse{
		ProductIds: recommended,
	}, nil
}
