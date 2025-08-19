package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	testpb "controlplaneplayground/testpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// server is used to implement test.TestServiceServer.
type server struct {
	testpb.UnimplementedTestServiceServer
	message string
}

// SayHello implements test.TestServiceServer
func (s *server) SayHello(ctx context.Context, in *testpb.HelloRequest) (*testpb.HelloReply, error) {
	log.Printf("Received request for: %v", in.GetName())
	return &testpb.HelloReply{Message: "Hello " + in.GetName() + " from " + s.message + " server"}, nil
}

func main() {
	var port string
	var message string
	var bindAll bool

	flag.StringVar(&port, "port", "50051", "Port to listen on")
	flag.StringVar(&message, "message", "test", "Server message identifier")
	flag.BoolVar(&bindAll, "bind-all", false, "Bind to 0.0.0.0 instead of 127.0.0.1 (required for Docker)")
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
	testpb.RegisterTestServiceServer(s, &server{message: message})

	// Register reflection service on gRPC server
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
