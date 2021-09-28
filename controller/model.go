package controller

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ControllerStatus struct {
	Replicas  int `json:"replicas"`
	Succeeded int `json:"succeeded"`
}

type ControllerSpec struct {
	Message string `json:"message"`
}

type SyncRequestChildren struct {
	Pods map[string]*corev1.Pod `json:"Pod.v1"`
}

type Controller struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ControllerSpec   `json:"spec"`
	Status            ControllerStatus `json:"status"`
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
