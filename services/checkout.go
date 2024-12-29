package services

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"

	"google.golang.org/grpc"

	"github.com/appnetorg/OnlineBoutique/protos/checkout"
)

// NewCheckoutService returns a new server for the CheckoutService
func NewCheckoutService(port int) *CheckoutService {
	return &CheckoutService{
		name: "checkout-service",
		port: port,
	}
}

// CheckoutService implements the CheckoutService
type CheckoutService struct {
	name string
	port int
	checkout.CheckoutServiceServer
}

// Run starts the server
func (s *CheckoutService) Run() error {
	srv := grpc.NewServer()
	checkout.RegisterCheckoutServiceServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("CheckoutService running at port: %d", s.port)
	return srv.Serve(lis)
}

// PlaceOrder processes an order placement request
func (s *CheckoutService) PlaceOrder(ctx context.Context, req *checkout.PlaceOrderRequest) (*checkout.PlaceOrderResponse, error) {
	log.Printf("PlaceOrder request received for user_id = %v", req.GetUserId())

	// Generate mock data for the response
	orderID := fmt.Sprintf("ORDER-%d", rand.Intn(1000000))
	trackingID := fmt.Sprintf("TRACKING-%d", rand.Intn(1000000))

	shippingCost := &checkout.Money{
		CurrencyCode: req.GetUserCurrency(),
		Units:        10,
		Nanos:        500000000,
	}

	orderItems := []*checkout.OrderItem{
		{
			Item: &checkout.CartItem{
				ProductId: "product-123",
				Quantity:  1,
			},
			Cost: &checkout.Money{
				CurrencyCode: req.GetUserCurrency(),
				Units:        50,
				Nanos:        0,
			},
		},
		{
			Item: &checkout.CartItem{
				ProductId: "product-456",
				Quantity:  2,
			},
			Cost: &checkout.Money{
				CurrencyCode: req.GetUserCurrency(),
				Units:        30,
				Nanos:        0,
			},
		},
	}

	response := &checkout.PlaceOrderResponse{
		Order: &checkout.OrderResult{
			OrderId:            orderID,
			ShippingTrackingId: trackingID,
			ShippingCost:       shippingCost,
			ShippingAddress:    req.GetAddress(),
			Items:              orderItems,
		},
	}

	return response, nil
}
