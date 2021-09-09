package main

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

func main() {
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/health", http.HandlerFunc(endpointHealth))
	http.Handle("/sync", http.HandlerFunc(endpointSync))
}

func endpointSync(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Sync is OK\n")
}

func endpointHealth(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("ok"))
}
