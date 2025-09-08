package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	testpb "controlplaneplayground/testpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type ServerConfig struct {
	Name       string `json:"name"`
	Protocol   string `json:"protocol"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Message    string `json:"message"`
	AlwaysFail bool   `json:"alwaysFail,omitempty"`
}

type Config struct {
	Servers []ServerConfig `json:"servers"`
}

// server is used to implement test.TestServiceServer.
type server struct {
	testpb.UnimplementedTestServiceServer
	message    string
	alwaysFail bool
}

// SayHello implements test.TestServiceServer
func (s *server) SayHello(ctx context.Context, in *testpb.HelloRequest) (*testpb.HelloReply, error) {
	log.Printf("Received gRPC request for: %v", in.GetName())
	if s.alwaysFail {
		return nil, status.Errorf(codes.Internal, "Server configured to always fail")
	}
	return &testpb.HelloReply{Message: "Hello " + in.GetName() + " from " + s.message + " server"}, nil
}

// HTTP handlers
type HelloRequest struct {
	Name string `json:"name"`
}

type HelloResponse struct {
	Message string `json:"message"`
}

func handleSayHello(w http.ResponseWriter, r *http.Request, message string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req HelloRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Received HTTP request for: %v", req.Name)

	response := HelloResponse{
		Message: "Hello " + req.Name + " from " + message + " server",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func startGRPCServer(config ServerConfig, wg *sync.WaitGroup) {
	defer wg.Done()

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", addr, err)
	}

	fmt.Printf("%s (gRPC) listening at %v\n", config.Name, lis.Addr())

	s := grpc.NewServer()
	testpb.RegisterTestServiceServer(s, &server{message: config.Message, alwaysFail: config.AlwaysFail})
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC on %s: %v", addr, err)
	}
}

func startHTTPServer(config ServerConfig, wg *sync.WaitGroup) {
	defer wg.Done()

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	http.HandleFunc("/test/sayhello", func(w http.ResponseWriter, r *http.Request) {
		handleSayHello(w, r, config.Message)
	})
	http.HandleFunc("/health", handleHealth)

	fmt.Printf("%s (HTTP) listening at %s\n", config.Name, addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("failed to serve HTTP on %s: %v", addr, err)
	}
}

func main() {
	var configFile string
	flag.StringVar(&configFile, "config", "config.json", "Path to config file")
	flag.Parse()

	// Read config file
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatalf("failed to parse config file: %v", err)
	}

	if len(config.Servers) == 0 {
		log.Fatal("no servers configured")
	}

	var wg sync.WaitGroup

	// Start each server in a separate goroutine
	for _, serverConfig := range config.Servers {
		wg.Add(1)
		switch serverConfig.Protocol {
		case "grpc":
			go startGRPCServer(serverConfig, &wg)
		case "http":
			go startHTTPServer(serverConfig, &wg)
		default:
			log.Fatalf("unsupported protocol: %s", serverConfig.Protocol)
		}
	}

	fmt.Printf("Started %d servers\n", len(config.Servers))

	// Wait for all servers to finish (they shouldn't unless there's an error)
	wg.Wait()
}
