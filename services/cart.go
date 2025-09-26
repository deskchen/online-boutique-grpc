package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"

	pb "github.com/deskchen/online-boutique-grpc/protos/onlineboutique"
)

// NewCartService returns a new server for the CartService
func NewCartService(port int, tracer opentracing.Tracer) *CartService {
	return &CartService{
		port:   port,
		Tracer: tracer,
	}
}

// CartService implements the CartService
type CartService struct {
	port int
	pb.CartServiceServer

	cartRedisAddr string
	rdb           *redis.Client // Redis client

	Tracer opentracing.Tracer
}

// Run starts the server
func (s *CartService) Run() error {

	mustMapEnv(&s.cartRedisAddr, "CART_REDIS_ADDR")

	s.rdb = redis.NewClient(&redis.Options{
		Addr: s.cartRedisAddr,
	})

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(s.Tracer)),
	}

	srv := grpc.NewServer(opts...)
	pb.RegisterCartServiceServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("CartService running at port: %d", s.port)
	return srv.Serve(lis)
}

// AddItem adds an item to the user's cart
func (s *CartService) AddItem(ctx context.Context, req *pb.AddItemRequest) (*pb.Empty, error) {
	log.Printf("AddItem request for user_id = %v, product_id = %v, quantity = %v", req.GetUserId(), req.GetItem().GetProductId(), req.GetItem().GetQuantity())

	userID := req.GetUserId()
	item := req.GetItem()

	// Fetch the existing cart
	data, err := s.rdb.Get(ctx, userID).Result()
	var cart []*pb.CartItem
	if err == redis.Nil {
		cart = []*pb.CartItem{} // Empty cart
	} else if err != nil {
		log.Printf("Failed to fetch cart for user_id = %v: %v", userID, err)
		return nil, err
	} else {
		err = json.Unmarshal([]byte(data), &cart)
		if err != nil {
			log.Printf("Failed to unmarshal cart for user_id = %v: %v", userID, err)
			return nil, err
		}
	}

	// Add item to the cart
	cart = append(cart, item)

	// Save the updated cart
	cartData, err := json.Marshal(cart)
	if err != nil {
		log.Printf("Failed to marshal cart for user_id = %v: %v", userID, err)
		return nil, err
	}

	err = s.rdb.Set(ctx, userID, cartData, 0).Err()
	if err != nil {
		log.Printf("Failed to save cart for user_id = %v: %v", userID, err)
		return nil, err
	}

	return &pb.Empty{}, nil
}

// GetCart retrieves the cart for a user
func (s *CartService) GetCart(ctx context.Context, req *pb.GetCartRequest) (*pb.Cart, error) {
	log.Printf("GetCart request for user_id = %v", req.GetUserId())

	userID := req.GetUserId()
	data, err := s.rdb.Get(ctx, userID).Result()
	if err == redis.Nil {
		return &pb.Cart{
			UserId: userID,
			Items:  []*pb.CartItem{},
		}, nil
	} else if err != nil {
		log.Printf("Failed to fetch cart for user_id = %v: %v", userID, err)
		return nil, err
	}

	var cart []*pb.CartItem
	err = json.Unmarshal([]byte(data), &cart)
	if err != nil {
		log.Printf("Failed to unmarshal cart for user_id = %v: %v", userID, err)
		return nil, err
	}

	return &pb.Cart{
		UserId: userID,
		Items:  cart,
	}, nil
}

// EmptyCart clears the cart for a user
func (s *CartService) EmptyCart(ctx context.Context, req *pb.EmptyCartRequest) (*pb.Empty, error) {
	log.Printf("EmptyCart request for user_id = %v", req.GetUserId())

	err := s.rdb.Del(ctx, req.GetUserId()).Err()
	if err != nil {
		log.Printf("Failed to delete cart for user_id = %v: %v", req.GetUserId(), err)
		return nil, err
	}

	return &pb.Empty{}, nil
}
