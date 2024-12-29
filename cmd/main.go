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
		// frontendport       = flag.Int("frontendport", 8080, "frontend service port")
		cartport           = flag.Int("cartaddr", 8081, "cart service port")
		productcatalogport = flag.Int("productcatalogport", 8082, "productcatalog service port")
		currencyport       = flag.Int("currencyport", 8083, "currency service port")
		paymentport        = flag.Int("paymentport", 8084, "payment service port")
		shippingport       = flag.Int("shippingport", 8085, "shipping service port")
		emailport          = flag.Int("emailport", 8086, "email service port")
		checkoutport       = flag.Int("checkoutport", 8087, "checkout service port")
		recommendationport = flag.Int("recommendationport", 8088, "recommendation service port")
		adport             = flag.Int("adport", 8089, "ad service port")

		// cartaddr           = flag.String("cartaddr", "cart:8080", "cart service addr")
		// productcatalogaddr = flag.String("productcatalogaddr", "productcatalog:8081", "productcatalog service addr")
		// currencyaddr       = flag.String("currencyaddr", "currency:8082", "currency service addr")
		// paymentaddr        = flag.String("paymentaddr", "payment:8083", "payment service addr")
		// shippingaddr       = flag.String("shippingaddr", "shipping:8084", "shipping service addr")
		// emailaddr          = flag.String("emailaddr", "email:8085", "email service addr")
		// checkoutaddr       = flag.String("checkoutaddr", "checkout:8086", "checkout service addr")
		// recommendationaddr = flag.String("recommendationaddr", "recommendation:8087", "recommendation service addr")
		// adaddr             = flag.String("adaddr", "ad:8088", "ad service addr")

		// cart_redis_addr = flag.String("cart_redis_addr", "redis:6379", "cart redis addr")
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
	default:
		log.Fatalf("unknown cmd: %s", cmd)
	}

	if err := srv.Run(); err != nil {
		log.Fatalf("run %s error: %v", cmd, err)
	}
}
