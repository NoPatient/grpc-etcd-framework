package main

import (
	"context"
	"google.golang.org/grpc"
	"grpc-etcd-framework/cfg"
	pb "grpc-etcd-framework/gen/go/v1"
	"grpc-etcd-framework/registry"
	"log"
	"net"
)

type server struct {
	pb.UnimplementedHelloServiceServer
}

func (*server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
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
	pb.RegisterHelloServiceServer(s, &server{})

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
