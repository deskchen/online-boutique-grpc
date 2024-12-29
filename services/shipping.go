package services

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"

	"google.golang.org/grpc"

	"github.com/appnetorg/OnlineBoutique/protos/shipping"
)

// NewShippingService returns a new server for the ShippingService
func NewShippingService(port int) *ShippingService {
	return &ShippingService{
		name: "shipping-service",
		port: port,
	}
}

// ShippingService implements the ShippingService
type ShippingService struct {
	name string
	port int
	shipping.ShippingServiceServer
}

// Run starts the server
func (s *ShippingService) Run() error {
	srv := grpc.NewServer()
	shipping.RegisterShippingServiceServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("ShippingService running at port: %d", s.port)
	return srv.Serve(lis)
}

// GetQuote calculates a shipping quote for a given address and items
func (s *ShippingService) GetQuote(ctx context.Context, req *shipping.GetQuoteRequest) (*shipping.GetQuoteResponse, error) {
	log.Printf("GetQuote request received for address: %v, %v, %v, %v, %v",
		req.GetAddress().GetStreetAddress(),
		req.GetAddress().GetCity(),
		req.GetAddress().GetState(),
		req.GetAddress().GetCountry(),
		req.GetAddress().GetZipCode())

	log.Printf("Calculating quote for %d items", len(req.GetItems()))

	// Mock quote calculation: $5 base + $2 per item
	cost := int64(5 + 2*len(req.GetItems()))

	response := &shipping.GetQuoteResponse{
		CostUsd: &shipping.Money{
			CurrencyCode: "USD",
			Units:        cost,
			Nanos:        0,
		},
	}

	return response, nil
}

// ShipOrder processes a shipping order and returns a tracking ID
func (s *ShippingService) ShipOrder(ctx context.Context, req *shipping.ShipOrderRequest) (*shipping.ShipOrderResponse, error) {
	log.Printf("ShipOrder request received for address: %v, %v, %v, %v, %v",
		req.GetAddress().GetStreetAddress(),
		req.GetAddress().GetCity(),
		req.GetAddress().GetState(),
		req.GetAddress().GetCountry(),
		req.GetAddress().GetZipCode())

	log.Printf("Shipping %d items", len(req.GetItems()))

	// Mock tracking ID generation
	trackingID := fmt.Sprintf("TRACKING-%d", rand.Intn(1000000))

	response := &shipping.ShipOrderResponse{
		TrackingId: trackingID,
	}

	log.Printf("Order shipped with tracking ID: %v", trackingID)

	return response, nil
}
