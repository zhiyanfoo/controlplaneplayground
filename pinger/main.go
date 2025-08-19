package main

import (
	"context"
	"log"
	"math/rand"
	"time"

	testpb "controlplaneplayground/testpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	log.Println("Starting pinger service...")

	// Connect to Envoy
	conn, err := grpc.Dial("localhost:10002", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Envoy: %v", err)
	}
	defer conn.Close()

	client := testpb.NewTestServiceClient(conn)

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Ping every 5 seconds
	ticker := time.NewTicker(time.Second/10)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			resp, err := client.SayHello(ctx, &testpb.HelloRequest{Name: "Pinger"})
			cancel()

			if err != nil {
				log.Printf("Ping failed: %v", err)
			} else {
				log.Printf("Ping successful: %s", resp.Message)
			}
		}
	}
}
