package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func endpointSync(writer http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodPost {
		writer.Header().Set("content-type", "application/json")
		writer.WriteHeader(http.StatusOK)
	} else {
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func endpointHealth(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("ok"))
}

func main() {
	log.Printf("Sandbox MetaController is about to start\n")
	log.Printf("Ku-Ku")
	address := "0.0.0.0:8000"

	http.Handle("/health", http.HandlerFunc(endpointHealth))
	http.Handle("/sync", http.HandlerFunc(endpointSync))

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
