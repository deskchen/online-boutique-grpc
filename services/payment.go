package services

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"

	"google.golang.org/grpc"

	"github.com/appnetorg/OnlineBoutique/protos/payment"
)

// NewPaymentService returns a new server for the PaymentService
func NewPaymentService(port int) *PaymentService {
	return &PaymentService{
		name: "payment-service",
		port: port,
	}
}

// PaymentService implements the PaymentService
type PaymentService struct {
	name string
	port int
	payment.PaymentServiceServer
}

// Run starts the server
func (s *PaymentService) Run() error {
	srv := grpc.NewServer()
	payment.RegisterPaymentServiceServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("PaymentService running at port: %d", s.port)
	return srv.Serve(lis)
}

// Charge processes a payment charge request
func (s *PaymentService) Charge(ctx context.Context, req *payment.ChargeRequest) (*payment.ChargeResponse, error) {
	log.Printf("Charge request received for amount: %v %v", req.GetAmount().GetCurrencyCode(), req.GetAmount().GetUnits())
	log.Printf("Credit Card Info: Number ending in ****%s, Expiry: %02d/%04d",
		req.GetCreditCard().GetCreditCardNumber()[len(req.GetCreditCard().GetCreditCardNumber())-4:],
		req.GetCreditCard().GetCreditCardExpirationMonth(),
		req.GetCreditCard().GetCreditCardExpirationYear())

	// Simulate transaction ID generation
	transactionID := fmt.Sprintf("TXN-%d", rand.Intn(1000000))

	log.Printf("Transaction successful: %v", transactionID)

	return &payment.ChargeResponse{
		TransactionId: transactionID,
	}, nil
}
