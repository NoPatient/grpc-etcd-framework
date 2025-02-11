package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"grpc-etcd-framework/cfg"
	pb "grpc-etcd-framework/gen/go/v1"
	"grpc-etcd-framework/registry"
	"log"
	"time"
)

func callSayHello(reg *registry.Registry, name string) {
	addrs, err := reg.Discover("MyService")
	if err != nil {
		log.Fatalf("failed to discover service: %v", err)
	}

	if len(addrs) == 0 {
		log.Fatalf("no available service instances")
	}

	conn, err := grpc.NewClient(
		addrs[0],
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

func main() {
	// Deprecated methods
	//conn, err := grpc.Dial("localhost:9999", grpc.WithInsecure(), grpc.WithBlock())
	etcdEndpoints := []string{cfg.ETCDEndpoint}
	fmt.Printf("etcdEndpoints: %v\n", etcdEndpoints)
	reg, err := registry.NewRegistry(etcdEndpoints, 50)
	if err != nil {
		log.Fatalf("failed to new registry: %v", err)
	}
	callSayHello(reg, "Harvey")
}
