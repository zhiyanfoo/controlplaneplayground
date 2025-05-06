package main

import (
	"context"
	"fmt"
	"log"
	"net"

	testpb "controlplaneplayground/testpb"

	"google.golang.org/grpc"
)

const (
	port = "127.0.0.1:50051"
)

// server is used to implement test.TestServiceServer.
type server struct {
	testpb.UnimplementedTestServiceServer
}

// SayHello implements test.TestServiceServer
func (s *server) SayHello(ctx context.Context, in *testpb.HelloRequest) (*testpb.HelloReply, error) {
	log.Printf("Received request for: %v", in.GetName())
	return &testpb.HelloReply{Message: "Hello " + in.GetName() + " from test server"}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fmt.Printf("Test server listening at %v\n", lis.Addr())
	s := grpc.NewServer()
	testpb.RegisterTestServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
