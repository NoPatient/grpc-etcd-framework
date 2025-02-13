package main

import (
	"context"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"grpc-etcd-framework/cfg"
	pb "grpc-etcd-framework/gen/go/v1"
	"grpc-etcd-framework/registry"
	"log"
	"net"
)

type server struct {
	pb.UnimplementedHelloServiceServer
	limiter *rate.Limiter
}

func newServer() *server {
	return &server{
		limiter: rate.NewLimiter(rate.Limit(cfg.ServiceLimiterThreshold), cfg.ServiceLimiterThreshold),
	}
}

func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	if !s.limiter.Allow() {
		return nil, status.Errorf(codes.ResourceExhausted, "rate limited")
	}
	return &pb.HelloResponse{
		Message: "Hello " + req.Name,
	}, nil
}

func startServer(reg *registry.Registry) {
	lis, err := net.Listen("tcp", ":9999")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterHelloServiceServer(s, newServer())

	log.Printf("server listening at %v", lis.Addr())

	go func() {
		if err := reg.Register("MyService", ":9999"); err != nil {
			log.Fatalf("failed to register service: %v", err)
		} else {
			log.Printf("service registered at %v", lis.Addr())
		}
	}()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to server: %v", err)
	}
}

func main() {
	etcdEndpoints := []string{cfg.ETCDEndpoint}
	reg, err := registry.NewRegistry(etcdEndpoints, 50)
	if err != nil {
		log.Fatalf("failed to new registry: %v", err)
	}
	startServer(reg)
}
