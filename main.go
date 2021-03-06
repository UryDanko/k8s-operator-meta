package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sandbox/controller"
	"syscall"
	"time"
)

const generation = "v1"

func sync(request *controller.SyncRequest) (*controller.SyncResponse, error) {
	response := &controller.SyncResponse{
		ResyncAfterSeconds: 1 * 60,
	}

	for _, pod := range request.Children.Pods {
		response.Status.Replicas++
		if pod.Status.Phase == corev1.PodRunning {
			response.Status.Succeeded++
		}
	}

	for i := 0; i < int(*request.Parent.Spec.Replicas); i++ {
		pod := corev1.Pod{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Pod",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("nginx%d", i),
				Labels: map[string]string{
					"app":        "nginx",
					"component":  "backend",
					"generation": generation,
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Image: "gcr.io/google_containers/nginx-slim:0.8",
					Name:  "nginx-sandbox",
					Ports: []corev1.ContainerPort{{
						ContainerPort: 80,
						Name:          "nginx-sandbox",
					}},
					ImagePullPolicy: "Always",
				},
				},
			},
		}
		response.Children = append(response.Children, &pod)
	}

	svc := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "nginx-sandbox",
			Labels: map[string]string{
				"component":  "backend",
				"generation": generation,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:     "web",
				Protocol: corev1.ProtocolTCP,
				Port:     80,
			}},
			Selector: map[string]string{
				"app": "nginx",
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
	response.Children = append(response.Children, &svc)

	prettyJSON, _ := json.MarshalIndent(response, "", "  ")
	fmt.Println(string(prettyJSON))
	return response, nil
}

func handlerHealth(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("ok"))
}

func handlerFinalize(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("Finalizing..."))
}

func handlerSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	request := &controller.SyncRequest{}
	if err := json.Unmarshal(body, request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	response, err := sync(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	body, err = json.Marshal(&response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func main() {
	log.Printf("Sandbox MetaController is about to start\n")
	address := "0.0.0.0:8000"

	http.Handle("/health", http.HandlerFunc(handlerHealth))
	http.Handle("/sync", http.HandlerFunc(handlerSync))
	http.Handle("/finalize", http.HandlerFunc(handlerFinalize))

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
