/*
Copyright 2021 The Kubernetes-CSI-Addons Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReclaimSpaceSpec defines the required parameters for ReclaimSpace
// operation.
type ReclaimSpaceSpec struct {
	// PVC represents PersistentVolumeClaim name.
	PVC string `json:"pvc,omitempty"`

	// PVC represents PersistentVolume name.
	PV string `json:"pv,omitempty"`
}

// OperationSpec defines the desired operation of CSIAdoonsJobSpec.
type OperationSpec struct {
	// ReclaimSpace repesents ReclaimSpace operation.
	ReclaimSpace ReclaimSpaceSpec `json:"reclaimSpace,omitempty"`
}

// OperationState defines the state of the operation.
type OperationState string

const (
	// Succeeded represents the Succeeded operation state.
	Succeeded OperationState = "Succeeded"

	// Failed represents the Failed operation state.
	Failed OperationState = "Failed"
)

// CSIAddonsJobSpec defines the desired state of CSIAddonsJob
type CSIAddonsJobSpec struct {
	// Operation represents the operation to be performed on the volume.
	// Current Supported operation is "ReclaimSapce".
	// +kubebuilder:validation:Required
	Operation OperationSpec `json:"operation"`

	// ActiveDeadlineSeconds specifies the duration in seconds relative to the
	// startTime that the operation might be active; value must be positive integer.
	// +optional
	ActiveDeadlineSeconds int64 `json:"activeDeadlineSeconds"`

	// BackOffLimit Specifies the number of retries before marking this operation failed.
	// +optional
	BackoffLimit int32 `json:"backOffLimit"`
}

// CSIAddonsJobStatus defines the observed state of CSIAddonsJob
type CSIAddonsJobStatus struct {
	// State defines the state of the operation.
	State OperationState `json:"state,omitempty"`

	// Message contains any message from the operation.
	Message string `json:"message,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:JSONPath=".metadata.namespace",name=namespace,type=string
//+kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date
//+kubebuilder:printcolumn:JSONPath=".spec.operation.reclaimSpace.pvc",name=pvcName,type=string
//+kubebuilder:printcolumn:JSONPath=".spec.operation.reclaimSpace.pv",name=pvName,type=string
//+kubebuilder:printcolumn:JSONPath=".status.state",name=state,type=string
//+kubebuilder:printcolumn:JSONPath=".status.message",name=message,type=string

// CSIAddonsJob is the Schema for the csiaddonsjobs API
type CSIAddonsJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CSIAddonsJobSpec   `json:"spec,omitempty"`
	Status CSIAddonsJobStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CSIAddonsJobList contains a list of CSIAddonsJob
type CSIAddonsJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CSIAddonsJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CSIAddonsJob{}, &CSIAddonsJobList{})
}
