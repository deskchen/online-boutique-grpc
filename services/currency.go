package services

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/appnetorg/OnlineBoutique/protos/currency"
)

// NewCurrencyService returns a new server for the CurrencyService
func NewCurrencyService(port int) *CurrencyService {
	return &CurrencyService{
		name:                "currency-service",
		port:                port,
		supportedCurrencies: []string{"USD", "EUR", "JPY", "GBP", "AUD"}, // Mock data
		exchangeRates: map[string]map[string]float64{
			"USD": {"EUR": 0.85, "JPY": 110.0, "GBP": 0.75, "AUD": 1.4},
			"EUR": {"USD": 1.18, "JPY": 130.0, "GBP": 0.88, "AUD": 1.65},
		}, // Mock exchange rates
	}
}

// CurrencyService implements the CurrencyService
type CurrencyService struct {
	name string
	port int
	currency.CurrencyServiceServer
	supportedCurrencies []string
	exchangeRates       map[string]map[string]float64
}

// Run starts the server
func (s *CurrencyService) Run() error {
	srv := grpc.NewServer()
	currency.RegisterCurrencyServiceServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("CurrencyService running at port: %d", s.port)
	return srv.Serve(lis)
}

// GetSupportedCurrencies returns a list of supported currency codes
func (s *CurrencyService) GetSupportedCurrencies(ctx context.Context, req *currency.Empty) (*currency.GetSupportedCurrenciesResponse, error) {
	log.Printf("GetSupportedCurrencies request received")
	return &currency.GetSupportedCurrenciesResponse{
		CurrencyCodes: s.supportedCurrencies,
	}, nil
}

// Convert converts an amount of money from one currency to another
func (s *CurrencyService) Convert(ctx context.Context, req *currency.CurrencyConversionRequest) (*currency.Money, error) {
	log.Printf("Convert request: from = %v %v, to = %v", req.GetFrom().GetUnits(), req.GetFrom().GetCurrencyCode(), req.GetToCode())

	from := req.GetFrom()
	toCode := req.GetToCode()

	rate, ok := s.exchangeRates[from.GetCurrencyCode()][toCode]
	if !ok {
		return nil, fmt.Errorf("unsupported currency conversion: %v to %v", from.GetCurrencyCode(), toCode)
	}

	// Perform conversion
	totalUnits := float64(from.GetUnits()) + float64(from.GetNanos())*1e-9
	convertedTotal := totalUnits * rate
	convertedUnits := int64(convertedTotal)
	convertedNanos := int32((convertedTotal - float64(convertedUnits)) * 1e9)

	return &currency.Money{
		CurrencyCode: toCode,
		Units:        convertedUnits,
		Nanos:        convertedNanos,
	}, nil
}
