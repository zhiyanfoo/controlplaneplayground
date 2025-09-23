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
	"google.golang.org/grpc/metadata"
)

func main() {
	var (
		addr          = flag.String("addr", "localhost:10000", "Address to connect to")
		interval      = flag.Duration("interval", time.Second, "Interval between requests")
		name          = flag.String("name", "Pinger", "Name to send in requests")
		sessionHeader = flag.String("session-header", "", "Session header name to track and send")
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

	// Track session header value if configured
	var sessionValue string

	if *sessionHeader != "" {
		log.Printf("Running continuously with %v interval, tracking session header '%s' (press Ctrl+C to stop)", *interval, *sessionHeader)
	} else {
		log.Printf("Running continuously with %v interval (press Ctrl+C to stop)", *interval)
	}

	for {
		select {
		case <-sigChan:
			log.Printf("Received shutdown signal. Total requests made: %d", requestCount)
			return

		case <-ticker.C:

			requestCount++
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

			// Add session header to request if we have a value
			if *sessionHeader != "" && sessionValue != "" {
				md := metadata.Pairs(*sessionHeader, sessionValue)
				ctx = metadata.NewOutgoingContext(ctx, md)
				log.Printf("[%d] Sending session header: %s=%s", requestCount, *sessionHeader, sessionValue)
			}

			// Capture response headers
			var respHeaders metadata.MD
			start := time.Now()
			resp, err := client.SayHello(ctx, &testpb.HelloRequest{Name: *name + " #" + strconv.Itoa(requestCount)}, grpc.Header(&respHeaders))
			duration := time.Since(start)
			cancel()

			if err != nil {
				log.Printf("[%d] Request failed (took %v): %v", requestCount, duration, err)
			} else {
				log.Printf("[%d] Request successful (took %v): %s", requestCount, duration, resp.Message)

				// Check for session header in response and store it
				if *sessionHeader != "" {
					if values := respHeaders.Get(*sessionHeader); len(values) > 0 {
						newSessionValue := values[0]
						if newSessionValue != sessionValue {
							log.Printf("[%d] Received new session header: %s=%s", requestCount, *sessionHeader, newSessionValue)
							sessionValue = newSessionValue
						} else {
							log.Printf("[%d] Session header unchanged: %s=%s", requestCount, *sessionHeader, sessionValue)
						}
					} else {
						log.Printf("[%d] No session header '%s' found in response", requestCount, *sessionHeader)
					}
				}
			}
		}
	}
}
