package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "grpc-etcd-framework/gen/go/v1"
	"log"
	"time"
)

func main() {
	// Deprecated methods
	//conn, err := grpc.Dial("localhost:9999", grpc.WithInsecure(), grpc.WithBlock())
	conn, err := grpc.NewClient(
		"localhost:9999",
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
		Name: "Harvey",
	})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.Message)
}
