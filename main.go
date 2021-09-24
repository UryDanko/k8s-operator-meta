package main

import (
	"context"
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type MetaRequest struct { // for now we assume that there is always only one parent
	//Q CompositeController
	Children map[string]interface{} `json:"children"`
	//Parent Sandbox `json:"parent"`
	//Controller CompositeController `json:"controller"`
}

type MetaResponse struct {
	// Set the delay (in seconds, as a float) before an optional, one-time, per-object resync.
	ResyncAfterSeconds float32 `json:"resync_after_seconds"`
	// A JSON object that will completely replace the status field within the parent object.
	Status corev1.PodStatus `json:"status"`
	// A list of JSON objects representing all the desired children for this parent object.
	Children []interface{} `json:"children"`
}

func endpointSync(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var metaRequest MetaRequest
		var metaResponse = MetaResponse{
			ResyncAfterSeconds: 2 * 60,
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{{
					Status:  corev1.ConditionTrue,
					Type:    corev1.PodReady,
					Reason:  "TestReason",
					Message: "TestMessage",
				}},
			},
		}

		//bodyBytes, err := ioutil.ReadAll(r.Body)
		//if err != nil {
		//	log.Fatal(err)
		//}
		//bodyString := string(bodyBytes)
		//fmt.Println(bodyString)

		if err := json.NewDecoder(r.Body).Decode(&metaRequest); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		childrenPods := metaRequest.Children["Pod.v1"]
		prettyJSON, _ := json.MarshalIndent(childrenPods, "", "  ")
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
					"type":       "corev1",
					"generation": "v4",
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

		metaResponse.Children = append(metaResponse.Children, pod)

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		prettyJSON, _ = json.MarshalIndent(metaResponse, "", "  ")
		fmt.Println(string(prettyJSON))
		json.NewEncoder(w).Encode(metaResponse)
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
