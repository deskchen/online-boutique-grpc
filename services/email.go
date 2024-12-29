package services

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/appnetorg/OnlineBoutique/protos/email"
)

// NewEmailService returns a new server for the EmailService
func NewEmailService(port int) *EmailService {
	return &EmailService{
		name: "email-service",
		port: port,
	}
}

// EmailService implements the EmailService
type EmailService struct {
	name string
	port int
	email.EmailServiceServer
}

// Run starts the server
func (s *EmailService) Run() error {
	srv := grpc.NewServer()
	email.RegisterEmailServiceServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("EmailService running at port: %d", s.port)
	return srv.Serve(lis)
}

// SendOrderConfirmation sends an order confirmation email
func (s *EmailService) SendOrderConfirmation(ctx context.Context, req *email.SendOrderConfirmationRequest) (*email.Empty, error) {
	log.Printf("SendOrderConfirmation request received for email = %v", req.GetEmail())

	order := req.GetOrder()
	log.Printf("Order ID: %v", order.GetOrderId())
	log.Printf("Shipping Tracking ID: %v", order.GetShippingTrackingId())
	log.Printf("Shipping Cost: %v %v", order.GetShippingCost().GetCurrencyCode(), order.GetShippingCost().GetUnits())
	log.Printf("Shipping Address: %v, %v, %v, %v, %v",
		order.GetShippingAddress().GetStreetAddress(),
		order.GetShippingAddress().GetCity(),
		order.GetShippingAddress().GetState(),
		order.GetShippingAddress().GetCountry(),
		order.GetShippingAddress().GetZipCode())

	for _, item := range order.GetItems() {
		log.Printf("Item: Product ID = %v, Quantity = %v, Cost = %v %v",
			item.GetItem().GetProductId(),
			item.GetItem().GetQuantity(),
			item.GetCost().GetCurrencyCode(),
			item.GetCost().GetUnits())
	}

	// Simulate sending an email (replace with actual email logic)
	log.Printf("Order confirmation email sent to %v", req.GetEmail())

	return &email.Empty{}, nil
}
