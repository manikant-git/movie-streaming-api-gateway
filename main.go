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

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("[API Gateway] Received request: %s %s\n", req.Method, req.URL.Path)

	// Check authentication
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

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := NewRouter()
	log.Printf("[API Gateway] Starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":" + port, router))
}
