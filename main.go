package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

type Router struct {
	routes map[string]string
}

func NewRouter() *Router {
	return &Router{
		routes: map[string]string{
			"/api/v1/movies":  "http://movie-service:8081",
			"/api/v1/users":   "http://user-service:8082",
			"/api/v1/stream":  "http://streaming-service:8083",
			"/api/v1/search":  "http://search-service:8084",
		},
	}
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("[API Gateway] Received request: %s %s\n", req.Method, req.URL.Path)

	// Check authentication (skip for health check)
	token := req.Header.Get("Authorization")
	if !strings.HasPrefix(token, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, `{"error": "Missing or invalid authorization token"}`)
		return
	}

	// Find matching service
	var target string
	for prefix, serviceURL := range r.routes {
		if strings.HasPrefix(req.URL.Path, prefix) {
			target = serviceURL
			break
		}
	}

	if target == "" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `{"error": "Service not found"}`)
		return
	}

	// Proxy request
	targetURL, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.ServeHTTP(w, req)
}

func healthHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "healthy", "service": "api-gateway"}`)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := NewRouter()

	// Health check endpoint (no auth required)
	http.HandleFunc("/health", healthHandler)

	// API routes (with auth)
	http.Handle("/api/", router)

	log.Printf("[API Gateway] Starting on port %s...\n", port)
	log.Printf("[API Gateway] Health check: GET /health\n")
	log.Printf("[API Gateway] API endpoints: /api/v1/{movies,users,stream,search}\n")
	log.Fatal(http.ListenAndServe(":" + port, nil))
}
