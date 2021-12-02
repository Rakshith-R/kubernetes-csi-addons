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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReclaimSpaceCronJobSpec defines the desired state of ReclaimSpaceCronJob
type ReclaimSpaceCronJobSpec struct {
	// The schedule in Cron format, see https://en.wikipedia.org/wiki/Cron.
	Schedule string `json:"schedule"`

	// This flag tells the controller to suspend subsequent executions, it does
	// not apply to already started executions. Defaults to false.
	// +optional
	// +kubebuilder:default:=false
	Suspend bool `json:"suspend"`

	// The number of finished job status history to retain,
	// Defaults to 3.
	// +optional
	// +kubebuilder:default:=3
	// +kubebuilder:validation:maximum:=100
	JobHistoryLimit int64 `json:"jobHistoryLimit"`

	// JobSpec contains information required to invoke reclaim space job
	JobSpec ReclaimSpaceJobSpec `json:"jobSpec"`
}

// ReclaimSpaceCronJobStatus defines the observed state of ReclaimSpaceCronJob
type ReclaimSpaceCronJobStatus struct {
	CurrentJobStatus   ReclaimSpaceJobStatus    `json:"currentJobStatus"`
	JobHistory         []*ReclaimSpaceJobStatus `json:"jobHistory"`
	LastStartTime      *metav1.Time             `json:"lastStartTime,omitempty"`
	LastCompletionTime *metav1.Time             `json:"lastCompletionTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ReclaimSpaceCronJob is the Schema for the reclaimspacecronjobs API
type ReclaimSpaceCronJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReclaimSpaceCronJobSpec   `json:"spec,omitempty"`
	Status ReclaimSpaceCronJobStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ReclaimSpaceCronJobList contains a list of ReclaimSpaceCronJob
type ReclaimSpaceCronJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ReclaimSpaceCronJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ReclaimSpaceCronJob{}, &ReclaimSpaceCronJobList{})
}
