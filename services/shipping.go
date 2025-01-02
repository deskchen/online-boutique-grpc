package services

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"

	"google.golang.org/grpc"

	pb "github.com/appnetorg/OnlineBoutique/protos/onlineboutique"
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
	pb.ShippingServiceServer
}

// Run starts the server
func (s *ShippingService) Run() error {
	srv := grpc.NewServer()
	pb.RegisterShippingServiceServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("ShippingService running at port: %d", s.port)
	return srv.Serve(lis)
}

// GetQuote calculates a shipping quote for a given address and items
func (s *ShippingService) GetQuote(ctx context.Context, req *pb.GetQuoteRequest) (*pb.GetQuoteResponse, error) {
	log.Printf("GetQuote request received for address: %v, %v, %v, %v, %v",
		req.GetAddress().GetStreetAddress(),
		req.GetAddress().GetCity(),
		req.GetAddress().GetState(),
		req.GetAddress().GetCountry(),
		req.GetAddress().GetZipCode())

	log.Printf("Calculating quote for %d items", len(req.GetItems()))

	// Generate a quote based on item count
	quote := createQuoteFromCount(len(req.GetItems()))

	response := &pb.GetQuoteResponse{
		CostUsd: &pb.Money{
			CurrencyCode: "USD",
			Units:        int64(quote.Dollars),
			Nanos:        int32(quote.Cents * 10000000),
		},
	}

	return response, nil
}

// ShipOrder processes a shipping order and returns a tracking ID
func (s *ShippingService) ShipOrder(ctx context.Context, req *pb.ShipOrderRequest) (*pb.ShipOrderResponse, error) {
	log.Printf("ShipOrder request received for address: %v, %v, %v, %v, %v",
		req.GetAddress().GetStreetAddress(),
		req.GetAddress().GetCity(),
		req.GetAddress().GetState(),
		req.GetAddress().GetCountry(),
		req.GetAddress().GetZipCode())

	log.Printf("Shipping %d items", len(req.GetItems()))

	// Generate tracking ID
	baseAddress := fmt.Sprintf("%s, %s, %s", req.GetAddress().GetStreetAddress(), req.GetAddress().GetCity(), req.GetAddress().GetState())
	trackingID := createTrackingID(baseAddress)

	response := &pb.ShipOrderResponse{
		TrackingId: trackingID,
	}

	log.Printf("Order shipped with tracking ID: %v", trackingID)

	return response, nil
}

// Quote represents a currency value.
type quote struct {
	Dollars uint32
	Cents   uint32
}

// createQuoteFromCount generates a shipping quote based on item count.
func createQuoteFromCount(count int) quote {
	return createQuoteFromFloat(8.99) // Example static rate
}

// createQuoteFromFloat generates a quote from a float value.
func createQuoteFromFloat(value float64) quote {
	units, fraction := math.Modf(value)
	return quote{
		uint32(units),
		uint32(math.Trunc(fraction * 100)),
	}
}

// createTrackingID generates a tracking ID.
func createTrackingID(salt string) string {
	return fmt.Sprintf("%c%c-%d%s-%d%s",
		getRandomLetterCode(),
		getRandomLetterCode(),
		len(salt),
		getRandomNumber(3),
		len(salt)/2,
		getRandomNumber(7),
	)
}

// getRandomLetterCode generates a code point value for a capital letter.
func getRandomLetterCode() uint32 {
	return 65 + uint32(rand.Intn(25))
}

// getRandomNumber generates a string representation of a number with the requested number of digits.
func getRandomNumber(digits int) string {
	str := ""
	for i := 0; i < digits; i++ {
		str = fmt.Sprintf("%s%d", str, rand.Intn(10))
	}
	return str
}
