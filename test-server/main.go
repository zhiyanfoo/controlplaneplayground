package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	testpb "controlplaneplayground/testpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// server is used to implement test.TestServiceServer.
type server struct {
	testpb.UnimplementedTestServiceServer
	message    string
	alwaysFail bool
}

// SayHello implements test.TestServiceServer
func (s *server) SayHello(ctx context.Context, in *testpb.HelloRequest) (*testpb.HelloReply, error) {
	log.Printf("Received request for: %v", in.GetName())
	if s.alwaysFail {
		return nil, status.Errorf(codes.Internal, "Server configured to always fail")
	}
	return &testpb.HelloReply{Message: "Hello " + in.GetName() + " from " + s.message + " server"}, nil
}

func main() {
	fmt.Printf("hello world 2\n")
	var port string
	var message string
	var bindAll bool
	var alwaysFail bool

	flag.StringVar(&port, "port", "50051", "Port to listen on")
	flag.StringVar(&message, "message", "test", "Server message identifier")
	flag.BoolVar(&bindAll, "bind-all", false, "Bind to 0.0.0.0 instead of 127.0.0.1 (required for Docker)")
	flag.BoolVar(&alwaysFail, "always-fail", false, "Always return errors for outlier detection testing")
	flag.Parse()

	// Determine bind address
	bindAddr := "127.0.0.1"
	if bindAll {
		bindAddr = "0.0.0.0"
	}

	// Construct full address
	addr := bindAddr + ":" + port

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fmt.Printf("Test server listening at %v\n", lis.Addr())
	s := grpc.NewServer()
	testpb.RegisterTestServiceServer(s, &server{message: message, alwaysFail: alwaysFail})

	// Register health service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, healthServer)
	
	// Set serving status for the service
	healthServer.SetServingStatus("my-test-server", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register reflection service on gRPC server
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
