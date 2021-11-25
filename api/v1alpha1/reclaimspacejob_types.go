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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OperationState defines the state of the operation.
type Result string

const (
	// ResultSucceeded represents the Succeeded operation state.
	ResultSucceeded Result = "Succeeded"

	// ResultFailed represents the Failed operation state.
	ResultFailed Result = "Failed"
)

// TargetSpec defines the targets on which the operation can be
// performed.
type TargetSpec struct {
	// PVC represents PersistentVolumeClaim name.
	PVC string `json:"pvc,omitempty"`
}

// ReclaimSpaceJobSpec defines the desired state of ReclaimSpaceJob
type ReclaimSpaceJobSpec struct {
	// Target represents volume target on which the operation will be
	// performed.
	// +kubebuilder:validation:Required
	Target TargetSpec `json:"target"`

	// BackOffLimit specifies the number of retries allowed before marking reclaim
	// space operation as failed. If not specified, defaults to 6.
	// +optional
	// +kubebuilder:default:=6
	BackoffLimit int32 `json:"backOffLimit"`

	// ActiveDeadlineSeconds specifies the duration in seconds relative to the
	// start time that the operation might be retried; value must be positive integer.
	// If not specified, defaults to 600 seconds.
	// +optional
	// +kubebuilder:default:=600
	ActiveDeadlineSeconds int64 `json:"activeDeadlineSeconds"`
}

// ReclaimSpaceJobStatus defines the observed state of ReclaimSpaceJob
type ReclaimSpaceJobStatus struct {
	// Result indicates the result of ReclaimSpaceJob.
	Result Result `json:"result,omitempty"`

	// Message contains any message from the ReclaimSpaceJob.
	Message string `json:"message,omitempty"`

	// ReclaimedSpace indicates the amount of space reclaimed.
	ReclaimedSpace resource.Quantity `json:"reclaimedSpace,omitempty"`

	// Conditions are the list of conditions and their status.
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObervedGeneration is the last generation change the controller has dealt with.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Retries indicates the number of times the operation is retried.
	Retries        int64        `json:"retries,omitempty"`
	StartTime      *metav1.Time `json:"lastStartTime,omitempty"`
	CompletionTime *metav1.Time `json:"lastCompletionTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:JSONPath=".metadata.namespace",name=Namespace,type=string
//+kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date
//+kubebuilder:printcolumn:JSONPath=".status.retries",name=Retries,type=integer
//+kubebuilder:printcolumn:JSONPath=".status.result",name=Result,type=string
//+kubebuilder:printcolumn:JSONPath=".status.reclaimedSpace",name=ReclaimedSpace,type=string,priority=1

// ReclaimSpaceJob is the Schema for the reclaimspacejobs API
type ReclaimSpaceJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReclaimSpaceJobSpec   `json:"spec,omitempty"`
	Status ReclaimSpaceJobStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ReclaimSpaceJobList contains a list of ReclaimSpaceJob
type ReclaimSpaceJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ReclaimSpaceJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ReclaimSpaceJob{}, &ReclaimSpaceJobList{})
}
