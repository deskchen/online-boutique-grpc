package services

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"

	"google.golang.org/grpc"

	"github.com/appnetorg/OnlineBoutique/protos/ad"
)

const (
	maxAdsToServe = 2
)

// NewAdService returns a new server for the AdService
func NewAdService(port int) *AdService {
	return &AdService{
		name: "ad-service",
		port: port,
		ads:  createAdsMap(),
	}
}

// AdService implements the AdService
type AdService struct {
	name string
	port int
	ads  map[string]*ad.Ad
	ad.AdServiceServer
}

// Run starts the server
func (s *AdService) Run() error {
	opts := []grpc.ServerOption{}
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

	var allAds []*ad.Ad
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

	return &ad.AdResponse{
		Ads: allAds,
	}, nil
}

func (s *AdService) getAdsByCategory(category string) []*ad.Ad {
	if adInstance, ok := s.ads[category]; ok {
		return []*ad.Ad{adInstance}
	}
	return nil
}

func (s *AdService) getRandomAds() []*ad.Ad {
	ads := make([]*ad.Ad, maxAdsToServe)
	vals := make([]*ad.Ad, 0, len(s.ads))
	for _, ad := range s.ads {
		vals = append(vals, ad)
	}
	for i := 0; i < maxAdsToServe; i++ {
		ads[i] = vals[rand.Intn(len(vals))]
	}
	return ads
}

func createAdsMap() map[string]*ad.Ad {
	return map[string]*ad.Ad{
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
