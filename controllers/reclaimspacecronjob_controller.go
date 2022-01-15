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

package controllers

import (
	"context"
	"fmt"
	"sort"
	"time"

	csiaddonsv1alpha1 "github.com/csi-addons/kubernetes-csi-addons/api/v1alpha1"

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

var (
	jobOwnerKey             = ".metadata.controller"
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

	var childJobs csiaddonsv1alpha1.ReclaimSpaceJobList
	if err := r.List(ctx, &childJobs, client.InNamespace(req.Namespace), client.MatchingFields{jobOwnerKey: req.Name}); err != nil {
		logger.Error(err, "Failed to list child ReclaimSpaceJobs")
		return ctrl.Result{}, err
	}

	// find the active list of jobs
	var activeJob *csiaddonsv1alpha1.ReclaimSpaceJob
	var successfulJobs, failedJobs []*csiaddonsv1alpha1.ReclaimSpaceJob
	var mostRecentTime *time.Time // find the last run so we can update the status

	for i, job := range childJobs.Items {
		switch job.Status.Result {
		case "": // ongoing
			activeJob = &childJobs.Items[i]
		case csiaddonsv1alpha1.OperationResultFailed:
			failedJobs = append(failedJobs, &childJobs.Items[i])
		case csiaddonsv1alpha1.OperationResultSucceeded:
			successfulJobs = append(successfulJobs, &childJobs.Items[i])
		}

		// We'll store the launch time in an annotation, so we'll reconstitute that from
		// the active jobs themselves.
		scheduledTimeForJob, err := getScheduledTimeForRSJob(&job)
		if err != nil {
			logger.Error(err, "Failed to parse schedule time for child ReclaimSpaceJob", "ReclaimSpaceJob", &job)
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
	rsCronJob.Status.Active = nil
	if activeJob != nil {
		jobRef, err := ref.GetReference(r.Scheme, activeJob)
		if err != nil {
			logger.Error(err, "Failed to make reference to active job", "job", activeJob)
		}
		rsCronJob.Status.Active = jobRef
	}

	logger.V(1).Info("job count", "successful jobs", len(successfulJobs), "failed jobs", len(failedJobs))

	if err := r.Status().Update(ctx, rsCronJob); err != nil {
		logger.Error(err, "unable to update status")
		return ctrl.Result{}, err
	}

	if rsCronJob.Spec.FailedJobsHistoryLimit != nil {
		sort.Slice(failedJobs, func(i, j int) bool {
			if failedJobs[i].Status.StartTime == nil {
				return failedJobs[j].Status.StartTime != nil
			}
			return failedJobs[i].Status.StartTime.Before(failedJobs[j].Status.StartTime)
		})
		for i, job := range failedJobs {
			if int32(i) >= int32(len(failedJobs))-*rsCronJob.Spec.FailedJobsHistoryLimit {
				break
			}
			if err := r.Delete(ctx, job, client.PropagationPolicy(metav1.DeletePropagationBackground)); client.IgnoreNotFound(err) != nil {
				logger.Error(err, "unable to delete old failed job", "job", job)
			} else {
				logger.V(0).Info("deleted old failed job", "job", job)
			}
		}
	}

	if rsCronJob.Spec.SuccessfulJobsHistoryLimit != nil {
		sort.Slice(successfulJobs, func(i, j int) bool {
			if successfulJobs[i].Status.StartTime == nil {
				return successfulJobs[j].Status.StartTime != nil
			}
			return successfulJobs[i].Status.StartTime.Before(successfulJobs[j].Status.StartTime)
		})
		for i, job := range successfulJobs {
			if int32(i) >= int32(len(successfulJobs))-*rsCronJob.Spec.SuccessfulJobsHistoryLimit {
				break
			}
			if err := r.Delete(ctx, job, client.PropagationPolicy(metav1.DeletePropagationBackground)); (err) != nil {
				logger.Error(err, "unable to delete old successful job", "job", job)
			} else {
				logger.V(0).Info("deleted old successful job", "job", job)
			}
		}
	}

	if rsCronJob.Spec.Suspend != nil && *rsCronJob.Spec.Suspend {
		logger.V(1).Info("rsCronJob suspended, skipping")
		return ctrl.Result{}, nil
	}

	// figure out the next times that we need to create
	// jobs at (or anything we missed).
	missedRun, nextRun, err := getNextSchedule(rsCronJob, r.Now())
	if err != nil {
		logger.Error(err, "unable to figure out CronJob schedule")
		// we don't really care about requeuing until we get an update that
		// fixes the schedule, so don't return an error
		return ctrl.Result{}, nil
	}

	/*
		We'll prep our eventual request to requeue until the next job, and then figure
		out if we actually need to run.
	*/
	scheduledResult := ctrl.Result{RequeueAfter: nextRun.Sub(r.Now())} // save this so we can re-use it elsewhere
	logger = logger.WithValues("now", time.Now(), "next run", nextRun)

	/*
		### 6: Run a new job if it's on schedule, not past the deadline, and not blocked by our concurrency policy
		If we've missed a run, and we're still within the deadline to start it, we'll need to run a job.
	*/
	if missedRun.IsZero() {
		logger.Info("no upcoming scheduled times, sleeping until next")
		return scheduledResult, nil
	}

	// make sure we're not too late to start the run
	logger = logger.WithValues("current run", missedRun)
	tooLate := false
	if rsCronJob.Spec.StartingDeadlineSeconds != nil {
		tooLate = missedRun.Add(time.Duration(*rsCronJob.Spec.StartingDeadlineSeconds) * time.Second).Before(r.Now())
	}
	if tooLate {
		logger.Info("missed starting deadline for last run, sleeping till next")
		return scheduledResult, nil
	}

	/*
		If we actually have to run a job, we'll need to either wait till existing ones finish,
		replace the existing ones, or just add new ones.  If our information is out of date due
		to cache delay, we'll get a requeue when we get up-to-date information.
	*/
	// figure out how to run this job
	// replace existing ones
	if rsCronJob.Spec.ConcurrencyPolicy == csiaddonsv1alpha1.ReplaceConcurrent {
		// we don't care if the job was already deleted
		err = r.Delete(ctx, activeJob, client.PropagationPolicy(metav1.DeletePropagationBackground))
		if client.IgnoreNotFound(err) != nil {
			logger.Error(err, "unable to delete active job", "job", activeJob)
			return ctrl.Result{}, err
		}
	}

	// default is to forbid concurrent execution
	if activeJob != nil {
		logger.V(1).Info("concurrency policy blocks concurrent runs, skipping", "active job", activeJob)
		return scheduledResult, nil
	}

	// actually make the job...f
	rsJob, err := r.constructRSJobForCronJob(rsCronJob, missedRun)
	if err != nil {
		logger.Error(err, "unable to construct job from template")
		// don't bother requeuing until we get a change to the spec
		return scheduledResult, nil
	}

	// ...and create it on the cluster
	if err := r.Create(ctx, rsJob); err != nil {
		logger.Error(err, "unable to create reclaimSpaceJob for reclaimSpaceCronJob", "reclaimSpaceJob", rsJob)
		return ctrl.Result{}, err
	}

	logger.V(1).Info("created reclaimSpaceJob for reclaimSpaceCronJob run", "reclaimSpacejob", rsJob)

	/*
		### 7: Requeue when we either see a running job or it's time for the next scheduled run
		Finally, we'll return the result that we prepped above, that says we want to requeue
		when our next run would need to occur.  This is taken as a maximum deadline -- if something
		else changes in between, like our job starts or finishes, we get modified, etc, we might
		reconcile again sooner.
	*/
	// we'll requeue once we see the running job, and update our status
	return scheduledResult, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReclaimSpaceCronJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &csiaddonsv1alpha1.ReclaimSpaceJob{}, jobOwnerKey, func(rawObj client.Object) []string {
		// grab the job object, extract the owner...
		job := rawObj.(*csiaddonsv1alpha1.ReclaimSpaceJob)
		owner := metav1.GetControllerOf(job)
		if owner == nil {
			return nil
		}
		// ...make sure it's a CronJob...
		if owner.APIVersion != apiGVStr || owner.Kind != "ReclaimSpaceCronJob" {
			return nil
		}

		// ...and if so, return it
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
func (r *ReclaimSpaceCronJobReconciler) constructRSJobForCronJob(rsCronJob *csiaddonsv1alpha1.ReclaimSpaceCronJob, scheduledTime time.Time) (*csiaddonsv1alpha1.ReclaimSpaceJob, error) {
	// We want job names for a given nominal start time to have a deterministic name to avoid the same job being created twice
	name := fmt.Sprintf("%s-%d", rsCronJob.Name, scheduledTime.Unix())

	job := &csiaddonsv1alpha1.ReclaimSpaceJob{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
			Name:        name,
			Namespace:   rsCronJob.Namespace,
		},
		Spec: *rsCronJob.Spec.ReclaimSpaceJobTemplate.Spec.DeepCopy(),
	}
	for k, v := range rsCronJob.Spec.ReclaimSpaceJobTemplate.Annotations {
		job.Annotations[k] = v
	}
	job.Annotations[scheduledTimeAnnotation] = scheduledTime.Format(time.RFC3339)
	for k, v := range rsCronJob.Spec.ReclaimSpaceJobTemplate.Labels {
		job.Labels[k] = v
	}
	if err := ctrl.SetControllerReference(rsCronJob, job, r.Scheme); err != nil {
		return nil, err
	}

	return job, nil
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

// getNextSchedule returns lastMissed and next time after parsing the schedule.
func getNextSchedule(rsCronJob *csiaddonsv1alpha1.ReclaimSpaceCronJob, now time.Time) (lastMissed time.Time, next time.Time, err error) {
	sched, err := cron.ParseStandard(rsCronJob.Spec.Schedule)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("Unparseable schedule %q: %v", rsCronJob.Spec.Schedule, err)
	}

	// for optimization purposes, cheat a bit and start from our last observed run time
	// we could reconstitute this here, but there's not much point, since we've
	// just updated it.
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
		// An object might miss several starts. For example, if
		// controller gets wedged on Friday at 5:01pm when everyone has
		// gone home, and someone comes in on Tuesday AM and discovers
		// the problem and restarts the controller, then all the hourly
		// jobs, more than 80 of them for one hourly scheduledJob, should
		// all start running with no further intervention (if the scheduledJob
		// allows concurrency and late starts).
		//
		// However, if there is a bug somewhere, or incorrect clock
		// on controller's server or apiservers (for setting creationTimestamp)
		// then there could be so many missed start times (it could be off
		// by decades or more), that it would eat up all the CPU and memory
		// of this controller. In that case, we want to not try to list
		// all the missed start times.
		starts++
		if starts > 100 {
			// We can't get the most recent times so just return an empty slice
			return time.Time{}, time.Time{}, fmt.Errorf("Too many missed start times (> 100). Set or decrease .spec.startingDeadlineSeconds or check clock skew.")
		}
	}
	return lastMissed, sched.Next(now), nil
}
