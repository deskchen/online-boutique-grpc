package services

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/appnetorg/OnlineBoutique/protos/cart"
)

// NewCartService returns a new server for the CartService
func NewCartService(port int) *CartService {
	return &CartService{
		name:  "cart-service",
		port:  port,
		carts: make(map[string][]*cart.CartItem), // In-memory storage for simplicity
	}
}

// CartService implements the CartService
type CartService struct {
	name string
	port int
	cart.CartServiceServer
	carts map[string][]*cart.CartItem // Mock in-memory storage for user carts
}

// Run starts the server
func (s *CartService) Run() error {
	srv := grpc.NewServer()
	cart.RegisterCartServiceServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("CartService running at port: %d", s.port)
	return srv.Serve(lis)
}

// AddItem adds an item to the user's cart
func (s *CartService) AddItem(ctx context.Context, req *cart.AddItemRequest) (*cart.Empty, error) {
	log.Printf("AddItem request for user_id = %v, product_id = %v, quantity = %v", req.GetUserId(), req.GetItem().GetProductId(), req.GetItem().GetQuantity())

	s.carts[req.GetUserId()] = append(s.carts[req.GetUserId()], req.GetItem())
	return &cart.Empty{}, nil
}

// GetCart retrieves the cart for a user
func (s *CartService) GetCart(ctx context.Context, req *cart.GetCartRequest) (*cart.Cart, error) {
	log.Printf("GetCart request for user_id = %v", req.GetUserId())

	items := s.carts[req.GetUserId()]
	return &cart.Cart{
		UserId: req.GetUserId(),
		Items:  items,
	}, nil
}

// EmptyCart clears the cart for a user
func (s *CartService) EmptyCart(ctx context.Context, req *cart.EmptyCartRequest) (*cart.Empty, error) {
	log.Printf("EmptyCart request for user_id = %v", req.GetUserId())

	delete(s.carts, req.GetUserId())
	return &cart.Empty{}, nil
}
