package main

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"grpc-etcd-framework/cfg"
	pb "grpc-etcd-framework/gen/go/v1"
	"grpc-etcd-framework/registry"
	"log"
	"net"
	"sync"
	"time"
)

type server struct {
	pb.UnimplementedHelloServiceServer
	limiter *rate.Limiter
	mu      sync.Mutex
}

func newServer() *server {
	return &server{
		limiter: rate.NewLimiter(rate.Limit(cfg.ServiceLimiterThreshold), cfg.ServiceLimiterThreshold),
	}
}

func (s *server) updateConfig(newConfig cfg.ServerConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.limiter.SetLimit(rate.Limit(newConfig.RateLimit))
}

func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	if !s.limiter.Allow() {
		return nil, status.Errorf(codes.ResourceExhausted, "rate limited")
	}
	return &pb.HelloResponse{
		Message: "Hello " + req.Name,
	}, nil
}

func startServer(reg *registry.Registry, configManager *cfg.ConfigManager) {
	lis, err := net.Listen("tcp", ":9999")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	selfServer := newServer()
	pb.RegisterHelloServiceServer(grpcServer, selfServer)

	log.Printf("server listening at %v", lis.Addr())

	go func() {
		if err := reg.Register("MyService", ":9999"); err != nil {
			log.Fatalf("failed to register service: %v", err)
		} else {
			log.Printf("service registered at %v", lis.Addr())
		}
	}()

	go func() {
		updates := make(chan cfg.ServerConfig)
		cli, err := clientv3.New(clientv3.Config{
			Endpoints:   []string{cfg.ETCDEndpoint},
			DialTimeout: 50 * time.Second,
		})
		if err != nil {
			log.Fatalf("failed to connect to etcd: %v", err)
		}
		defer cli.Close()

		go cfg.WatchConfigChanges(context.Background(), cli, cfg.ETCDConfigPath, updates)
		for newConfig := range updates {
			configManager.UpdateConfig(newConfig)
			selfServer.updateConfig(newConfig)
		}
	}()

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to server: %v", err)
	}
}

func main() {
	etcdEndpoints := []string{cfg.ETCDEndpoint}
	reg, err := registry.NewRegistry(etcdEndpoints, 50)
	if err != nil {
		log.Fatalf("failed to new registry: %v", err)
	}

	initialConfig := cfg.ServerConfig{
		RateLimit: cfg.ServiceLimiterThreshold,
	}
	configManager := cfg.NewConfigManager(initialConfig)

	startServer(reg, configManager)
}
