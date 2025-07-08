package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
)

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
	log.Printf("Received HTTP health request")
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func main() {
	var port string
	var message string

	flag.StringVar(&port, "port", "50052", "Port to listen on")
	flag.StringVar(&message, "message", "HTTP", "Server message identifier")
	flag.Parse()

	// Construct full address
	addr := "127.0.0.1:" + port

	// Create handler functions with message parameter
	http.HandleFunc("/test/sayhello", func(w http.ResponseWriter, r *http.Request) {
		handleSayHello(w, r, message)
	})
	http.HandleFunc("/health", handleHealth)

	fmt.Printf("HTTP test server listening at %v\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
