package main

import (
	"flag"
	"log"
	"os"

	services "github.com/appnetorg/OnlineBoutique/services"
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
	)
	flag.Parse()

	var srv server
	var cmd = os.Args[1]
	println("cmd parsed: ", cmd)

	switch cmd {
	case "cart":
		srv = services.NewCartService(*cartport)
	case "productcatalog":
		srv = services.NewProductCatalogService(*productcatalogport)
	case "currency":
		srv = services.NewCurrencyService(*currencyport)
	case "payment":
		srv = services.NewPaymentService(*paymentport)
	case "shipping":
		srv = services.NewShippingService(*shippingport)
	case "email":
		srv = services.NewEmailService(*emailport)
	case "checkout":
		srv = services.NewCheckoutService(*checkoutport)
	case "recommendation":
		srv = services.NewRecommendationService(*recommendationport)
	case "ad":
		srv = services.NewAdService(*adport)
	case "frontend":
		srv = services.NewFrontendServer(*frontendport)
	default:
		log.Fatalf("unknown cmd: %s", cmd)
	}

	if err := srv.Run(); err != nil {
		log.Fatalf("run %s error: %v", cmd, err)
	}
}
