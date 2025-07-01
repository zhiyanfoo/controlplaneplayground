package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const (
	httpPort = "127.0.0.1:50052"
)

type HelloRequest struct {
	Name string `json:"name"`
}

type HelloResponse struct {
	Message string `json:"message"`
}

func handleSayHello(w http.ResponseWriter, r *http.Request) {
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
		Message: "Hello " + req.Name + " from HTTP test server",
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

func main() {
	http.HandleFunc("/test/sayhello", handleSayHello)
	http.HandleFunc("/health", handleHealth)

	fmt.Printf("HTTP test server listening at %v\n", httpPort)
	log.Fatal(http.ListenAndServe(httpPort, nil))
}
