package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Controller struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ControllerSpec   `json:"spec"`
	Status            ControllerStatus `json:"status"`
}

type ControllerSpec struct {
	Message string `json:"message"`
}

type ControllerStatus struct {
	Replicas  int `json:"replicas"`
	Succeeded int `json:"succeeded"`
}

type SyncRequestChildren struct {
	Pods map[string]*corev1.Pod `json:"Pod.v1"`
}

type SyncRequest struct {
	Children SyncRequestChildren `json:"children"`
	Parent   Controller          `json:"parent"`
}

type SyncResponse struct {
	// Set the delay (in seconds, as a float) before an optional, one-time, per-object resync.
	ResyncAfterSeconds float32 `json:"resync_after_seconds"`
	// A JSON object that will completely replace the status field within the parent object.
	Status ControllerStatus `json:"status"`
	// A list of JSON objects representing all the desired children for this parent object.
	Children []runtime.Object `json:"children"`
}

func sync(request *SyncRequest) (*SyncResponse, error) {
	var response = &SyncResponse{
		ResyncAfterSeconds: 1 * 60,
	}

	for _, pod := range request.Children.Pods {
		response.Status.Replicas++
		if pod.Status.Phase == corev1.PodSucceeded {
			response.Status.Succeeded++
		}
	}

	prettyJSON, _ := json.MarshalIndent(request.Children, "", "  ")
	fmt.Println(string(prettyJSON))
	//if len(parentPods) == 0 {
	//	log.Printf("Parent POD list is empty!!!")
	//}

	pod := corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "nginx1",
			Labels: map[string]string{
				"app":        "nginx",
				"component":  "backend",
				"generation": "v2",
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

	prettyJSON, _ = json.MarshalIndent(response, "", "  ")
	fmt.Println(string(prettyJSON))
	return response, nil
}

func endpointHealth(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("ok"))
}

func endpointFinalize(writer http.ResponseWriter, request *http.Request) {
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

	//bodyBytes, err := ioutil.ReadAll(r.Body)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//bodyString := string(bodyBytes)
	//fmt.Println(bodyString)

	request := &SyncRequest{}
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

	http.Handle("/health", http.HandlerFunc(endpointHealth))
	http.Handle("/sync", http.HandlerFunc(handlerSync))
	http.Handle("/finalize", http.HandlerFunc(endpointFinalize))

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
