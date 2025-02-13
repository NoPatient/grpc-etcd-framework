package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"grpc-etcd-framework/cfg"
	"grpc-etcd-framework/circuitbreaker"
	pb "grpc-etcd-framework/gen/go/v1"
	"grpc-etcd-framework/registry"
	"log"
	"time"
)

func callSayHello(serviceAddrs []string, name string) {
	conn, err := grpc.NewClient(
		serviceAddrs[0],
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewHelloServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.SayHello(ctx, &pb.HelloRequest{
		Name: name,
	})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.Message)
}

func callSayHelloWithCircuitBreaker(serviceAddrs []string, name string, cb *circuitbreaker.CircuitBreaker) {
	conn, err := grpc.NewClient(
		serviceAddrs[0],
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
	callSayHello(serviceAddrs, "Harvey")

	cb := circuitbreaker.NewCircuitBreaker(5, 10*time.Second)
	for i := 0; i < 20; i++ {
		callSayHelloWithCircuitBreaker(serviceAddrs, "Harvey Breaker", cb)
		time.Sleep(100 * time.Millisecond)
	}
}
