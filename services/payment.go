package services

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	pb "github.com/appnetorg/OnlineBoutique/protos/onlineboutique"
)

type InvalidCreditCardErr struct{}

func (e InvalidCreditCardErr) Error() string {
	return "invalid credit card"
}

type UnacceptedCreditCardErr struct{}

func (e UnacceptedCreditCardErr) Error() string {
	return "credit card not accepted; only VISA or MasterCard are accepted"
}

type ExpiredCreditCardErr struct{}

func (e ExpiredCreditCardErr) Error() string {
	return "credit card expired"
}

func validateAndCharge(amount *pb.Money, card *pb.CreditCardInfo) (string, error) {
	// Perform some rudimentary validation.
	number := strings.ReplaceAll(card.CreditCardNumber, "-", "")
	var company string
	switch {
	case len(number) < 4:
		return "", InvalidCreditCardErr{}
	case number[0] == '4':
		company = "Visa"
	case number[0] == '5':
		company = "MasterCard"
	default:
		return "", UnacceptedCreditCardErr{}
	}

	if card.CreditCardCvv < 100 || card.CreditCardCvv > 9999 {
		return "", InvalidCreditCardErr{}
	}

	if time.Date(int(card.CreditCardExpirationYear), time.Month(card.CreditCardExpirationMonth), 0, 0, 0, 0, 0, time.Local).Before(time.Now()) {
		return "", ExpiredCreditCardErr{}
	}

	// Card is valid: process the transaction.
	log.Printf(
		"Transaction processed: company=%s, last_four=%s, currency=%s, amount=%d.%d",
		company,
		number[len(number)-4:],
		amount.CurrencyCode,
		amount.Units,
		amount.Nanos,
	)

	// Generate a transaction ID.
	return uuid.New().String(), nil
}

// NewPaymentService returns a new server for the PaymentService
func NewPaymentService(port int) *PaymentService {
	return &PaymentService{
		port: port,
	}
}

// PaymentService implements the PaymentService
type PaymentService struct {
	port int
	pb.PaymentServiceServer
}

// Run starts the server
func (s *PaymentService) Run() error {
	srv := grpc.NewServer()
	pb.RegisterPaymentServiceServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("PaymentService running at port: %d", s.port)
	return srv.Serve(lis)
}

// Charge processes a payment charge request
func (s *PaymentService) Charge(ctx context.Context, req *pb.ChargeRequest) (*pb.ChargeResponse, error) {
	log.Printf("Charge request received for amount: %v %v", req.GetAmount().GetCurrencyCode(), req.GetAmount().GetUnits())
	log.Printf("Credit Card Info: Number ending in ****%s, Expiry: %02d/%04d",
		req.GetCreditCard().GetCreditCardNumber()[len(req.GetCreditCard().GetCreditCardNumber())-4:],
		req.GetCreditCard().GetCreditCardExpirationMonth(),
		req.GetCreditCard().GetCreditCardExpirationYear())

	transactionID, err := validateAndCharge(req.GetAmount(), req.GetCreditCard())
	if err != nil {
		log.Printf("Transaction failed: %v", err)
		return nil, err
	}

	log.Printf("Transaction successful: %v", transactionID)

	return &pb.ChargeResponse{
		TransactionId: transactionID,
	}, nil
}
