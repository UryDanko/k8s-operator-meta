package main

import (
	"context"
	api "k8s.io/api/core/v1"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type MetaRequest struct {
	Pod api.Pod `json:"pod:object"`
}

type MetaResponse struct {
	Pod api.Pod `json:"pod:object"`
}

func endpointSync(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func endpointHealth(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("ok"))
}

func endpointFinalize(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("Finalizing..."))
}

func main() {
	log.Printf("Sandbox MetaController is about to start\n")
	address := "0.0.0.0:8000"

	http.Handle("/health", http.HandlerFunc(endpointHealth))
	http.Handle("/sync", http.HandlerFunc(endpointSync))
	http.Handle("/finalize", http.HandlerFunc(endpointFinalize))

	log.Printf("Starting....")
	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	server := &http.Server{
		Addr: address,
	}

	e := make(chan error)

	go func() {
		e <- server.ListenAndServe()
	}()

	select {
	case err := <-e:
		log.Fatalf("%v\n", err)
	case <-stop:
	}

	log.Printf("Received signal, gracefully shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Failed to shutdown: %v\n", err)
	}
}
