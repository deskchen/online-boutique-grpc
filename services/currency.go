package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net"
	"strconv"

	"google.golang.org/grpc"

	pb "github.com/appnetorg/OnlineBoutique/protos/onlineboutique"
)

const (
	filePath = "currency_conversion.json"
)

// CurrencyService implements the CurrencyService
type CurrencyService struct {
	name string
	port int
	pb.CurrencyServiceServer
	conversionMap map[string]float64
}

// NewCurrencyService returns a new server for the CurrencyService
func NewCurrencyService(port int) *CurrencyService {
	// Read the file content into a []byte
	currencyData, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	conversionMap, err := createConversionMap(currencyData)
	if err != nil {
		return nil
	}
	return &CurrencyService{
		name:          "currency-service",
		port:          port,
		conversionMap: conversionMap,
	}
}

// Run starts the server
func (s *CurrencyService) Run() error {
	srv := grpc.NewServer()
	pb.RegisterCurrencyServiceServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("CurrencyService running at port: %d", s.port)
	return srv.Serve(lis)
}

// GetSupportedCurrencies returns a list of supported currency codes
func (s *CurrencyService) GetSupportedCurrencies(ctx context.Context, req *pb.Empty) (*pb.GetSupportedCurrenciesResponse, error) {
	log.Printf("GetSupportedCurrencies request received")
	keys := make([]string, 0, len(s.conversionMap))
	for k := range s.conversionMap {
		keys = append(keys, k)
	}
	return &pb.GetSupportedCurrenciesResponse{
		CurrencyCodes: keys,
	}, nil
}

// Convert converts an amount of money from one currency to another
func (s *CurrencyService) Convert(ctx context.Context, req *pb.CurrencyConversionRequest) (*pb.Money, error) {
	log.Printf("Convert request: from = %v %v, to = %v", req.GetFrom().GetUnits(), req.GetFrom().GetCurrencyCode(), req.GetToCode())

	from := req.GetFrom()
	toCode := req.GetToCode()

	// Convert: from -> EUR
	fromRate, ok := s.conversionMap[from.GetCurrencyCode()]
	if !ok {
		return nil, fmt.Errorf("unsupported currency code: %v", from.GetCurrencyCode())
	}
	euros := carry(float64(from.GetUnits())/fromRate, float64(from.GetNanos())/fromRate)

	// Convert: EUR -> toCode
	toRate, ok := s.conversionMap[toCode]
	if !ok {
		return nil, fmt.Errorf("unsupported currency code: %v", toCode)
	}
	to := carry(float64(euros.Units)*toRate, float64(euros.Nanos)*toRate)
	to.CurrencyCode = toCode

	return &pb.Money{
		CurrencyCode: to.CurrencyCode,
		Units:        to.Units,
		Nanos:        to.Nanos,
	}, nil
}

// carry handles decimal/fractional carrying for currency conversions.
func carry(units float64, nanos float64) pb.Money {
	const fractionSize = 1000000000 // 1B
	nanos += math.Mod(units, 1.0) * fractionSize
	units = math.Floor(units) + math.Floor(nanos/fractionSize)
	nanos = math.Mod(nanos, fractionSize)
	return pb.Money{
		Units: int64(units),
		Nanos: int32(nanos),
	}
}

// createConversionMap parses the currency conversion JSON data.
func createConversionMap(currencyData []byte) (map[string]float64, error) {
	m := map[string]string{}
	if err := json.Unmarshal(currencyData, &m); err != nil {
		return nil, err
	}
	conv := make(map[string]float64, len(m))
	for k, v := range m {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, err
		}
		conv[k] = f
	}
	return conv, nil
}
