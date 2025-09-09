package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	testpb "controlplaneplayground/testpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	var (
		addr     = flag.String("addr", "localhost:10000", "Address to connect to")
		interval = flag.Duration("interval", time.Second, "Interval between requests")
		name     = flag.String("name", "Pinger", "Name to send in requests")
	)
	flag.Parse()

	log.Printf("Starting pinger client connecting to %s...", *addr)

	// Connect to Envoy
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to %s: %v", *addr, err)
	}
	defer conn.Close()

	client := testpb.NewTestServiceClient(conn)

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	requestCount := 0
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()
	log.Printf("Running continuously with %v interval (press Ctrl+C to stop)", *interval)

	for {
		select {
		case <-sigChan:
			log.Printf("Received shutdown signal. Total requests made: %d", requestCount)
			return

		case <-ticker.C:

			requestCount++
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			
			start := time.Now()
			resp, err := client.SayHello(ctx, &testpb.HelloRequest{Name: *name + " #" + strconv.Itoa(requestCount)})
			duration := time.Since(start)
			cancel()

			if err != nil {
				log.Printf("[%d] Request failed (took %v): %v", requestCount, duration, err)
			} else {
				log.Printf("[%d] Request successful (took %v): %s", requestCount, duration, resp.Message)
			}
		}
	}
}
