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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConcurrencyPolicy describes how the job will be handled.
// Only one of the following concurrent policies may be specified.
// If none of the following policies is specified, the default one
// is ForbidConcurrent.
// +enum
type ConcurrencyPolicy string

const (
	// ForbidConcurrent forbids concurrent runs, skipping next run if previous
	// hasn't finished yet.
	ForbidConcurrent ConcurrencyPolicy = "Forbid"

	// ReplaceConcurrent cancels currently running job and replaces it with a new one.
	ReplaceConcurrent ConcurrencyPolicy = "Replace"
)

// ReclaimSpaceJobTemplateSpec describes the data a Job should have when created from a template
type ReclaimSpaceJobTemplateSpec struct {
	// Standard object's metadata of the jobs created from this template.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specification of the desired behavior of the job.
	// +optional
	Spec ReclaimSpaceJobSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// ReclaimSpaceCronJobSpec defines the desired state of ReclaimSpaceCronJob
type ReclaimSpaceCronJobSpec struct {

	// The schedule in Cron format, see https://en.wikipedia.org/wiki/Cron.
	Schedule string `json:"schedule"`

	// Optional deadline in seconds for starting the job if it misses scheduled
	// time for any reason.  Missed jobs executions will be counted as failed ones.
	// +optional
	StartingDeadlineSeconds *int64 `json:"startingDeadlineSeconds,omitempty"`

	// Specifies how to treat concurrent executions of a Job.
	// Valid values are:
	// - "Forbid" (default): forbids concurrent runs, skipping next run if previous run hasn't finished yet;
	// - "Replace": cancels currently running job and replaces it with a new one
	// +optional
	ConcurrencyPolicy ConcurrencyPolicy `json:"concurrencyPolicy,omitempty"`

	// This flag tells the controller to suspend subsequent executions, it does
	// not apply to already started executions.  Defaults to false.
	// +optional
	Suspend *bool `json:"suspend,omitempty"`

	// Specifies the job that will be created when executing a CronJob.
	ReclaimSpaceJobTemplate ReclaimSpaceJobTemplateSpec `json:"reclaimSpaceJobTemplate"`

	// The number of successful finished jobs to retain. Value must be non-negative integer.
	// Defaults to 3.
	// +optional
	SuccessfulJobsHistoryLimit *int32 `json:"successfulJobsHistoryLimit,omitempty"`

	// The number of failed finished jobs to retain. Value must be non-negative integer.
	// Defaults to 1.
	// +optional
	FailedJobsHistoryLimit *int32 `json:"failedJobsHistoryLimit,omitempty"`
}

// ReclaimSpaceCronJobStatus defines the observed state of ReclaimSpaceCronJob
type ReclaimSpaceCronJobStatus struct {
	// A list of pointers to currently running jobs.
	// +optional
	Active *v1.ObjectReference `json:"active,omitempty"`

	// Information when was the last time the job was successfully scheduled.
	// +optional
	LastScheduleTime *metav1.Time `json:"lastScheduleTime,omitempty"`

	// Information when was the last time the job successfully completed.
	// +optional
	LastSuccessfulTime *metav1.Time `json:"lastSuccessfulTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:JSONPath=".spec.schedule",name=Schedule,type=string
//+kubebuilder:printcolumn:JSONPath=".spec.suspend",name=Suspend,type=boolean
//+kubebuilder:printcolumn:JSONPath=".status.active.name",name=Active,type=string
//+kubebuilder:printcolumn:JSONPath=".status.lastScheduleTime",name=Lastschedule,type=date
//+kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date
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
