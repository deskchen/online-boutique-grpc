package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "github.com/deskchen/online-boutique-grpc/protos/onlineboutique"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
)

// Embed the HTML template for the email
var (
	tmpl = template.Must(template.New("email").
		Funcs(template.FuncMap{
			"div": func(x, y int32) int32 { return x / y },
		}).
		Parse("./templates/email.html"))
)

// NewEmailService returns a new server for the EmailService
func NewEmailService(port int) *EmailService {
	return &EmailService{
		port: port,
	}
}

// EmailService implements the EmailService
type EmailService struct {
	port int
	pb.EmailServiceServer
}

// Run starts the server
func (s *EmailService) Run() error {
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(opentracing.GlobalTracer())),
	}
	srv := grpc.NewServer(opts...)
	pb.RegisterEmailServiceServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("EmailService running at port: %d", s.port)
	return srv.Serve(lis)
}

// SendOrderConfirmation sends an order confirmation email
func (s *EmailService) SendOrderConfirmation(ctx context.Context, req *pb.SendOrderConfirmationRequest) (*pb.Empty, error) {
	log.Printf("SendOrderConfirmation request received for email = %v", req.GetEmail())

	// Generate email content using the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, req.GetOrder()); err != nil {
		log.Printf("Error executing template: %v", err)
		return nil, err
	}
	confirmation := buf.String()

	// Simulate sending the email
	log.Printf("Order confirmation email content for %v:\n%s", req.GetEmail(), confirmation)

	// Replace this with actual email-sending logic if needed
	log.Printf("Order confirmation email sent to %v", req.GetEmail())

	return &pb.Empty{}, nil
}
