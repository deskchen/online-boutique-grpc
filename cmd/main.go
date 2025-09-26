package main

import (
	"flag"
	"log"
	"os"

	"strconv"

	services "github.com/deskchen/online-boutique-grpc/services"
	"github.com/deskchen/online-boutique-grpc/services/tracing"
)

type server interface {
	Run() error
}

func main() {
	var (
		// port            = flag.Int("port", 8080, "The service port")
		frontendport       = flag.Int("frontendport", 8080, "frontend service port")
		cartport           = flag.Int("cartaddr", 8081, "cart service port")
		productcatalogport = flag.Int("productcatalogport", 8082, "productcatalog service port")
		currencyport       = flag.Int("currencyport", 8083, "currency service port")
		paymentport        = flag.Int("paymentport", 8084, "payment service port")
		shippingport       = flag.Int("shippingport", 8085, "shipping service port")
		emailport          = flag.Int("emailport", 8086, "email service port")
		checkoutport       = flag.Int("checkoutport", 8087, "checkout service port")
		recommendationport = flag.Int("recommendationport", 8088, "recommendation service port")
		adport             = flag.Int("adport", 8089, "ad service port")
		jaegerport         = flag.Int("jaegerport", 6831, "jaeger agent port")
	)
	flag.Parse()

	var srv server
	var cmd = os.Args[1]
	println("cmd parsed: ", cmd)
	tracer, err := tracing.Init(cmd, "jaeger:"+strconv.Itoa(*jaegerport))
	if err != nil {
		log.Fatalf("ERROR: cannot init Jaeger: %v\n", err)
	}
	log.Printf("Jaeger Tracer Initialised for %s", cmd)

	switch cmd {
	case "cart":
		srv = services.NewCartService(*cartport, tracer)
	case "productcatalog":
		srv = services.NewProductCatalogService(*productcatalogport, tracer)
	case "currency":
		srv = services.NewCurrencyService(*currencyport, tracer)
	case "payment":
		srv = services.NewPaymentService(*paymentport, tracer)
	case "shipping":
		srv = services.NewShippingService(*shippingport, tracer)
	case "email":
		srv = services.NewEmailService(*emailport, tracer)
	case "checkout":
		srv = services.NewCheckoutService(*checkoutport, tracer)
	case "recommendation":
		srv = services.NewRecommendationService(*recommendationport, tracer)
	case "ad":
		srv = services.NewAdService(*adport, tracer)
	case "frontend":
		srv = services.NewFrontendServer(*frontendport, tracer)
	default:
		log.Fatalf("unknown cmd: %s", cmd)
	}

	if err := srv.Run(); err != nil {
		log.Fatalf("run %s error: %v", cmd, err)
	}
}
