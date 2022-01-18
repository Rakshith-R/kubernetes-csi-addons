/*
Copyright 2022 The Kubernetes-CSI-Addons Authors.

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

package controllers

import (
	"context"
	"fmt"
	"sort"
	"time"

	csiaddonsv1alpha1 "github.com/csi-addons/kubernetes-csi-addons/api/v1alpha1"
	"github.com/go-logr/logr"

	"github.com/robfig/cron/v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ref "k8s.io/client-go/tools/reference"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ReclaimSpaceCronJobReconciler reconciles a ReclaimSpaceCronJob object
type ReclaimSpaceCronJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	jobOwnerKey                             = ".metadata.controller"
	defaultFailedJobsHistoryLimit     int32 = 1
	defaultSuccessfulJobsHistoryLimit int32 = 3
)

var (
	apiGVStr                = csiaddonsv1alpha1.GroupVersion.String()
	scheduledTimeAnnotation = csiaddonsv1alpha1.GroupVersion.Group + "/scheduled-at"
)

//+kubebuilder:rbac:groups=csiaddons.openshift.io,resources=reclaimspacecronjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=csiaddons.openshift.io,resources=reclaimspacecronjobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=csiaddons.openshift.io,resources=reclaimspacecronjobs/finalizers,verbs=update
//+kubebuilder:rbac:groups=csiaddons.openshift.io,resources=reclaimspacejobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=csiaddons.openshift.io,resources=reclaimspacejobs/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ReclaimSpaceCronJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	logger := log.FromContext(ctx)

	// Fetch ReclaimSpaceCronJob instance
	rsCronJob := &csiaddonsv1alpha1.ReclaimSpaceCronJob{}
	err := r.Client.Get(ctx, req.NamespacedName, rsCronJob)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			logger.Info("ReclaimSpaceCronJob resource not found")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// set history limit defaults, if not specified.
	if rsCronJob.Spec.FailedJobsHistoryLimit == nil {
		*rsCronJob.Spec.FailedJobsHistoryLimit = defaultFailedJobsHistoryLimit
	}
	if rsCronJob.Spec.SuccessfulJobsHistoryLimit == nil {
		*rsCronJob.Spec.SuccessfulJobsHistoryLimit = defaultSuccessfulJobsHistoryLimit
	}

	var childJobs csiaddonsv1alpha1.ReclaimSpaceJobList
	if err := r.List(ctx, &childJobs, client.InNamespace(req.Namespace), client.MatchingFields{jobOwnerKey: req.Name}); err != nil {
		logger.Error(err, "Failed to list child ReclaimSpaceJobs")
		return ctrl.Result{}, err
	}

	// find the active,failed and successful list of jobs
	var activeJob *csiaddonsv1alpha1.ReclaimSpaceJob
	var successfulJobs, failedJobs []*csiaddonsv1alpha1.ReclaimSpaceJob
	var mostRecentTime, lastSuccessfulTime *time.Time // find the last run so we can update the status

	for i, job := range childJobs.Items {
		switch job.Status.Result {
		case "": // ongoing
			activeJob = &childJobs.Items[i]
		case csiaddonsv1alpha1.OperationResultFailed:
			failedJobs = append(failedJobs, &childJobs.Items[i])
		case csiaddonsv1alpha1.OperationResultSucceeded:
			successfulJobs = append(successfulJobs, &childJobs.Items[i])
			completionTime := &childJobs.Items[i].Status.CompletionTime.Time
			if lastSuccessfulTime != nil {
				lastSuccessfulTime = completionTime
			} else if completionTime != nil && lastSuccessfulTime.Before(*completionTime) {
				lastSuccessfulTime = completionTime
			}
		}

		// reconstitute scheduled time from annotation.
		scheduledTimeForJob, err := getScheduledTimeForRSJob(&job)
		if err != nil {
			logger.Error(err, "Failed to parse schedule time for child ReclaimSpaceJob", "ReclaimSpaceJob", job.Name)
			continue
		}
		if scheduledTimeForJob != nil {
			if mostRecentTime == nil {
				mostRecentTime = scheduledTimeForJob
			} else if mostRecentTime.Before(*scheduledTimeForJob) {
				mostRecentTime = scheduledTimeForJob
			}
		}
	}

	if mostRecentTime != nil {
		rsCronJob.Status.LastScheduleTime = &metav1.Time{Time: *mostRecentTime}
	} else {
		rsCronJob.Status.LastScheduleTime = nil
	}
	rsCronJob.Status.LastSuccessfulTime = &metav1.Time{Time: *lastSuccessfulTime}
	rsCronJob.Status.Active = nil
	if activeJob != nil {
		jobRef, err := ref.GetReference(r.Scheme, activeJob)
		if err != nil {
			logger.Error(err, "Failed to make reference to active job", "job", activeJob)
		}
		rsCronJob.Status.Active = jobRef
	}

	logger.Info("job count", "successfulJobs", len(successfulJobs), "failedJobs", len(failedJobs))

	if err := r.Status().Update(ctx, rsCronJob); err != nil {
		logger.Error(err, "unable to update status")
		return ctrl.Result{}, err
	}

	// delete jobs older than history limit.
	r.deleteOldJobs(ctx, &logger, successfulJobs, *rsCronJob.Spec.SuccessfulJobsHistoryLimit)
	r.deleteOldJobs(ctx, &logger, failedJobs, *rsCronJob.Spec.FailedJobsHistoryLimit)

	if rsCronJob.Spec.Suspend != nil && *rsCronJob.Spec.Suspend {
		logger.Info("ReclaimspaceCronJob suspended, skipping scheduling job")
		return ctrl.Result{}, nil
	}

	// figure out the next times that we need to create jobs at (or anything we missed).
	missedRun, nextRun, err := getNextSchedule(rsCronJob, r.Now())
	if err != nil {
		logger.Error(err, "Failed to Parse out CronJob schedule", "schedule", rsCronJob.Spec.Schedule)
		// invalid schedule, do not requeue.
		return ctrl.Result{}, nil
	}

	//	We'll prep our eventual request to requeue until the next job, and then figure
	//	out if we actually need to run.
	scheduledResult := ctrl.Result{RequeueAfter: nextRun.Sub(r.Now())} // save this so we can re-use it elsewhere
	logger = logger.WithValues("now", time.Now(), "nextRun", nextRun)

	// If we've missed a run, and we're still within the deadline to start it, we'll need to run a job.
	if missedRun.IsZero() {
		logger.Info("no upcoming scheduled times, sleeping until next")
		return scheduledResult, nil
	}

	// make sure we're not too late to start the run
	logger = logger.WithValues("currentRun", missedRun)
	tooLate := false
	if rsCronJob.Spec.StartingDeadlineSeconds != nil {
		tooLate = missedRun.Add(time.Duration(*rsCronJob.Spec.StartingDeadlineSeconds) * time.Second).Before(r.Now())
	}
	if tooLate {
		logger.Info("Missed starting deadline for last run, sleeping till next")
		return scheduledResult, nil
	}

	// replace existing ones, if replace concurrent policy is set.
	if rsCronJob.Spec.ConcurrencyPolicy == csiaddonsv1alpha1.ReplaceConcurrent {
		err = r.Delete(ctx, activeJob, client.PropagationPolicy(metav1.DeletePropagationBackground))
		if client.IgnoreNotFound(err) != nil {
			logger.Error(err, "unable to delete active job", "job", activeJob)
			return ctrl.Result{}, err
		}
	}

	// default is to forbid concurrent execution
	if activeJob != nil {
		logger.Info("Concurrency policy blocks concurrent runs, skipping", "activeJob", activeJob.Name)
		return scheduledResult, nil
	}

	rsJob, err := r.constructRSJobForCronJob(rsCronJob, missedRun)
	if err != nil {
		logger.Error(err, "Failed to construct job from template")
		// don't requeuing until we get a change to the spec
		return scheduledResult, nil
	}
	if err := r.Create(ctx, rsJob); err != nil {
		logger.Error(err, "Failed to create reclaimSpaceJob for reclaimSpaceCronJob")
		return ctrl.Result{}, err
	}
	logger.Info("Successfully created reclaimSpaceJob for reclaimSpaceCronJob run", "reclaimSpacejob", rsJob.Name)

	// Reconcile will be triggered if job starts or finishes, cronjob is modified, etc.
	// Requeue till next run.
	return scheduledResult, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReclaimSpaceCronJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &csiaddonsv1alpha1.ReclaimSpaceJob{}, jobOwnerKey, func(rawObj client.Object) []string {
		// extract the owner from job object.
		job := rawObj.(*csiaddonsv1alpha1.ReclaimSpaceJob)
		owner := metav1.GetControllerOf(job)
		if owner == nil {
			return nil
		}
		if owner.APIVersion != apiGVStr || owner.Kind != "ReclaimSpaceCronJob" {
			return nil
		}

		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&csiaddonsv1alpha1.ReclaimSpaceCronJob{}).
		Owns(&csiaddonsv1alpha1.ReclaimSpaceJob{}).
		Complete(r)
}

// Now returns the current local time.
func (r *ReclaimSpaceCronJobReconciler) Now() time.Time {
	return time.Now()
}

// constructRSJobForCronJob contructs reclaimspacejob.
func (r *ReclaimSpaceCronJobReconciler) constructRSJobForCronJob(
	rsCronJob *csiaddonsv1alpha1.ReclaimSpaceCronJob,
	scheduledTime time.Time) (*csiaddonsv1alpha1.ReclaimSpaceJob, error) {
	// We want job names for a given nominal start time to have a deterministic name to avoid the same job being created twice
	name := fmt.Sprintf("%s-%d", rsCronJob.Name, scheduledTime.Unix())

	job := &csiaddonsv1alpha1.ReclaimSpaceJob{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
			Name:        name,
			Namespace:   rsCronJob.Namespace,
		},
		Spec: *rsCronJob.Spec.JobSpec.Spec.DeepCopy(),
	}
	for k, v := range rsCronJob.Spec.JobSpec.Annotations {
		job.Annotations[k] = v
	}
	job.Annotations[scheduledTimeAnnotation] = scheduledTime.Format(time.RFC3339)
	for k, v := range rsCronJob.Spec.JobSpec.Labels {
		job.Labels[k] = v
	}
	if err := ctrl.SetControllerReference(rsCronJob, job, r.Scheme); err != nil {
		return nil, err
	}

	return job, nil
}

// deleteOldJobs sorts given jobList by StartTime and deletes jobs older than given historyLimit.
// Errors if any are only logged.
func (r *ReclaimSpaceCronJobReconciler) deleteOldJobs(
	ctx context.Context,
	logger *logr.Logger,
	jobsList []*csiaddonsv1alpha1.ReclaimSpaceJob,
	historyLimit int32) {

	sort.Slice(jobsList, func(i, j int) bool {
		if jobsList[i].Status.StartTime == nil {
			return jobsList[j].Status.StartTime != nil
		}
		return jobsList[i].Status.StartTime.Before(jobsList[j].Status.StartTime)
	})

	for i, job := range jobsList {
		if int32(i) >= int32(len(jobsList))-historyLimit {
			break
		}
		err := r.Delete(ctx, job, client.PropagationPolicy(metav1.DeletePropagationBackground))
		if client.IgnoreNotFound(err) != nil {
			logger.Error(err, "Failed to delete old job",
				"reclaimSpaceJobName", job.Name,
				"state", job.Status.Result)
		} else {
			logger.Info("Successfully deleted old job",
				"reclaimSpaceJobName", job.Name,
				"state", job.Status.Result)
		}
	}
}

// getScheduledTimeForRSJob extract the scheduled time from the annotation that
// is added during job creation.
func getScheduledTimeForRSJob(rsJob *csiaddonsv1alpha1.ReclaimSpaceJob) (*time.Time, error) {
	timeRaw := rsJob.Annotations[scheduledTimeAnnotation]
	if len(timeRaw) == 0 {
		return nil, nil
	}

	timeParsed, err := time.Parse(time.RFC3339, timeRaw)
	if err != nil {
		return nil, err
	}
	return &timeParsed, nil
}

// TODO: add unit test for getNextSchedule
// getNextSchedule returns lastMissed and next time after parsing the schedule.
// This function returns error if there are more than 100 missed start times.
func getNextSchedule(
	rsCronJob *csiaddonsv1alpha1.ReclaimSpaceCronJob,
	now time.Time) (lastMissed time.Time, next time.Time, err error) {
	sched, err := cron.ParseStandard(rsCronJob.Spec.Schedule)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("Unparseable schedule %q: %v", rsCronJob.Spec.Schedule, err)
	}

	var earliestTime time.Time
	if rsCronJob.Status.LastScheduleTime != nil {
		earliestTime = rsCronJob.Status.LastScheduleTime.Time
	} else {
		earliestTime = rsCronJob.ObjectMeta.CreationTimestamp.Time
	}
	if rsCronJob.Spec.StartingDeadlineSeconds != nil {
		// controller is not going to schedule anything below this point
		schedulingDeadline := now.Add(-time.Second * time.Duration(*rsCronJob.Spec.StartingDeadlineSeconds))

		if schedulingDeadline.After(earliestTime) {
			earliestTime = schedulingDeadline
		}
	}
	if earliestTime.After(now) {
		return time.Time{}, sched.Next(now), nil
	}

	starts := 0
	for t := sched.Next(earliestTime); !t.After(now); t = sched.Next(t) {
		lastMissed = t
		starts++
		if starts > 100 {
			// We can't get the most recent times so just return an empty slice
			return time.Time{}, time.Time{},
				fmt.Errorf("Too many missed start times (> 100). Set or decrease" +
					".spec.startingDeadlineSeconds or check clock skew.")
		}
	}
	return lastMissed, sched.Next(now), nil
}
