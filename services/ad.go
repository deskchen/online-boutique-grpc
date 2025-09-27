package services

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"

	pb "github.com/deskchen/online-boutique-grpc/protos/onlineboutique"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

const (
	maxAdsToServe = 2
)

// NewAdService returns a new server for the AdService
func NewAdService(port int) *AdService {
	return &AdService{
		port: port,
		ads:  createAdsMap(),
	}
}

// AdService implements the AdService
type AdService struct {
	port int
	ads  map[string]*pb.Ad
	pb.AdServiceServer
}

// Run starts the server
func (s *AdService) Run() error {
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(opentracing.GlobalTracer())),
	}
	srv := grpc.NewServer(opts...)
	pb.RegisterAdServiceServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("AdService running at port: %d", s.port)
	return srv.Serve(lis)
}

// GetAds returns a list of ads based on the context keys
func (s *AdService) GetAds(ctx context.Context, req *pb.AdRequest) (*pb.AdResponse, error) {
	log.Printf("GetAds request with context_keys = %v", req.GetContextKeys())

	var allAds []*pb.Ad
	keywords := req.GetContextKeys()

	if len(keywords) > 0 {
		for _, kw := range keywords {
			allAds = append(allAds, s.getAdsByCategory(kw)...)
		}
		if len(allAds) == 0 {
			// Serve random ads
			allAds = s.getRandomAds()
		}
	} else {
		allAds = s.getRandomAds()
	}

	return &pb.AdResponse{
		Ads: allAds,
	}, nil
}

func (s *AdService) getAdsByCategory(category string) []*pb.Ad {
	if adInstance, ok := s.ads[category]; ok {
		return []*pb.Ad{adInstance}
	}
	return nil
}

func (s *AdService) getRandomAds() []*pb.Ad {
	ads := make([]*pb.Ad, maxAdsToServe)
	vals := make([]*pb.Ad, 0, len(s.ads))
	for _, ad := range s.ads {
		vals = append(vals, ad)
	}
	for i := 0; i < maxAdsToServe; i++ {
		ads[i] = vals[rand.Intn(len(vals))]
	}
	return ads
}

func createAdsMap() map[string]*pb.Ad {
	return map[string]*pb.Ad{
		"hair": {
			RedirectUrl: "/product/2ZYFJ3GM2N",
			Text:        "Hairdryer for sale. 50% off.",
		},
		"clothing": {
			RedirectUrl: "/product/66VCHSJNUP",
			Text:        "Tank top for sale. 20% off.",
		},
		"accessories": {
			RedirectUrl: "/product/1YMWWN1N4O",
			Text:        "Watch for sale. Buy one, get second kit for free",
		},
		"footwear": {
			RedirectUrl: "/product/L9ECAV7KIM",
			Text:        "Loafers for sale. Buy one, get second one for free",
		},
		"decor": {
			RedirectUrl: "/product/0PUK6V6EV0",
			Text:        "Candle holder for sale. 30% off.",
		},
		"kitchen": {
			RedirectUrl: "/product/9SIQT8TOJO",
			Text:        "Bamboo glass jar for sale. 10% off.",
		},
	}
}
