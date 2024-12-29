package main

import (
	"flag"
	"log"
	"os"
)

type server interface {
	Run() error
}

func main() {
	var (
		// port            = flag.Int("port", 8080, "The service port")
		cartport           = flag.Int("cartaddr", 8080, "cart service port")
		productcatalogport = flag.Int("productcatalogport", 8081, "productcatalog service port")
		currencyport       = flag.Int("currencyport", 8082, "currency service port")
		paymentport        = flag.Int("paymentport", 8083, "payment service port")
		shippingport       = flag.Int("shippingport", 8084, "shipping service port")
		emailport          = flag.Int("emailport", 8085, "email service port")
		checkoutport       = flag.Int("checkoutport", 8086, "checkout service port")
		recommendationport = flag.Int("recommendationport", 8087, "recommendation service port")
		adport             = flag.Int("adport", 8088, "ad service port")

		cartaddr           = flag.String("cartaddr", "cart:8080", "cart service addr")
		productcatalogaddr = flag.String("productcatalogaddr", "productcatalog:8081", "productcatalog service addr")
		currencyaddr       = flag.String("currencyaddr", "currency:8082", "currency service addr")
		paymentaddr        = flag.String("paymentaddr", "payment:8083", "payment service addr")
		shippingaddr       = flag.String("shippingaddr", "shipping:8084", "shipping service addr")
		emailaddr          = flag.String("emailaddr", "email:8085", "email service addr")
		checkoutaddr       = flag.String("checkoutaddr", "checkout:8086", "checkout service addr")
		recommendationaddr = flag.String("recommendationaddr", "recommendation:8087", "recommendation service addr")
		adaddr             = flag.String("adaddr", "ad:8088", "ad service addr")

		cart_redis_addr = flag.String("cart_redis_addr", "redis:6379", "cart redis addr")
	)
	flag.Parse()

	var srv server
	var cmd = os.Args[1]
	println("cmd parsed: ", cmd)

	tracer, err := tracing.Init(cmd, *jaegeraddr)
	if err != nil {
		log.Fatalf("Got error while initializing jaeger agent for cmd %s: %v", cmd, err)
	}
	log.Printf("tracer inited for cmd %s", cmd)

	switch cmd {
	case "details":
		srv = services.NewDetails(
			*detailsport,
			tracer,
			*details_mongodb_addr,
		)
	case "ratings":
		srv = services.NewRatings(
			*ratingsport,
			tracer,
			*ratings_mongodb_addr,
		)
	case "reviews":
		srv = services.NewReviews(
			*reviewsport,
			*ratingsaddr,
			tracer,
			*reviews_mongodb_addr,
		)
	case "productpage":
		srv = services.NewProductPage(
			*productpageport,
			*reviewsaddr,
			*detailsaddr,
			tracer,
		)
	default:
		log.Fatalf("unknown cmd: %s", cmd)
	}

	if err := srv.Run(); err != nil {
		log.Fatalf("run %s error: %v", cmd, err)
	}
}
