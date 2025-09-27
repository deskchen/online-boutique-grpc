package services

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"

	"google.golang.org/grpc"

	pb "github.com/deskchen/online-boutique-grpc/protos/onlineboutique"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
)

// NewRecommendationService returns a new server for the RecommendationService
func NewRecommendationService(port int) *RecommendationService {
	return &RecommendationService{
		port: port,
	}
}

// RecommendationService implements the RecommendationService
type RecommendationService struct {
	port int

	pb.RecommendationServiceServer

	productCatalogSvcAddr string
	productCatalogSvcConn *grpc.ClientConn
}

// Run starts the server
func (s *RecommendationService) Run() error {
	ctx := context.Background()

	mustMapEnv(&s.productCatalogSvcAddr, "PRODUCT_CATALOG_SERVICE_ADDR")
	mustConnGRPC(ctx, &s.productCatalogSvcConn, s.productCatalogSvcAddr)

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(opentracing.GlobalTracer())),
	}
	srv := grpc.NewServer(opts...)
	pb.RegisterRecommendationServiceServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("RecommendationService running at port: %d", s.port)
	return srv.Serve(lis)
}

// ListRecommendations provides a list of recommended product IDs based on user and product history
func (s *RecommendationService) ListRecommendations(ctx context.Context, req *pb.ListRecommendationsRequest) (*pb.ListRecommendationsResponse, error) {
	log.Printf("ListRecommendations request received for user_id = %v, product_ids = %v", req.GetUserId(), req.GetProductIds())

	// Fetch a list of products from the product catalog.
	catalogProducts, err := pb.NewProductCatalogServiceClient(s.productCatalogSvcConn).ListProducts(ctx, &pb.EmptyUser{UserId: req.GetUserId()})
	if err != nil {
		log.Printf("Error fetching catalog products: %v", err)
		return nil, err
	}

	// Remove user-provided products from the catalog to avoid recommending them.
	userProductIDs := req.GetProductIds()
	userIDs := make(map[string]struct{}, len(userProductIDs))
	for _, id := range userProductIDs {
		userIDs[id] = struct{}{}
	}

	filtered := make([]string, 0, len(catalogProducts.Products))
	for _, product := range catalogProducts.Products {
		if _, ok := userIDs[product.Id]; !ok {
			filtered = append(filtered, product.Id)
		}
	}

	// Sample from filtered products and return them.
	rand.Shuffle(len(filtered), func(i, j int) { filtered[i], filtered[j] = filtered[j], filtered[i] })

	const maxResponses = 5
	recommended := filtered
	if len(filtered) > maxResponses {
		recommended = filtered[:maxResponses]
	}

	return &pb.ListRecommendationsResponse{
		ProductIds: recommended,
	}, nil
}
