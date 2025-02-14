package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"grpc-etcd-framework/balancer"
	"grpc-etcd-framework/cfg"
	"grpc-etcd-framework/circuitbreaker"
	pb "grpc-etcd-framework/gen/go/v1"
	"grpc-etcd-framework/registry"
	"log"
	"time"
)

func callSayHelloWithCircuitBreaker(name string, cb *circuitbreaker.CircuitBreaker, clientBalancer balancer.Balancer) {
	targetAddress, err := clientBalancer.Next(name)
	if err != nil {
		log.Fatalf("client balancer next failed: %v", err)
	}

	conn, err := grpc.NewClient(
		targetAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewHelloServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if !cb.Allow() {
		log.Printf("Request locked by circuit breaker")
		return
	}
	r, err := client.SayHello(ctx, &pb.HelloRequest{
		Name: name,
	})
	if err != nil {
		cb.Failure()
		log.Printf("Call SyaHello failed: %v", err)
	} else {
		cb.Success()
		log.Printf("Call SayHello success: %s", r.Message)
	}
}

func getServiceAddresses() []string {
	etcdEndpoints := []string{cfg.ETCDEndpoint}
	log.Printf("etcdEndpoints: %v", etcdEndpoints)

	reg, err := registry.NewRegistry(etcdEndpoints, 50)
	if err != nil {
		log.Fatalf("failed to new registry: %v", err)
	}

	addrs, err := reg.Discover(cfg.ServiceName)
	if err != nil {
		log.Fatalf("failed to discover service:	%v", err)
	}

	if len(addrs) == 0 {
		log.Fatalf("no available service instances")
	}

	return addrs
}

func main() {
	serviceAddrs := getServiceAddresses()
	clientBalancer := balancer.NewConsistentHashBalancer(serviceAddrs)

	cb := circuitbreaker.NewCircuitBreaker(5, 10*time.Second)
	for i := 0; i < 20; i++ {
		callSayHelloWithCircuitBreaker("Harvey Breaker", cb, clientBalancer)
		time.Sleep(100 * time.Millisecond)
	}
}
