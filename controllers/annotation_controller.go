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

	csiaddonsv1alpha1 "github.com/csi-addons/kubernetes-csi-addons/api/v1alpha1"
	"github.com/go-logr/logr"

	"github.com/robfig/cron/v3"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// AnnontationReconciler reconciles annotations on persistentVolumeClaim objects.
type AnnotationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var (
	rsCronJobScheduleTimeAnnotation = "reclaimspace." + csiaddonsv1alpha1.GroupVersion.Group + "/schedule"
	rsCronJobNameAnnotation         = "reclaimspace." + csiaddonsv1alpha1.GroupVersion.Group + "/cronjob"
)

const (
	defaultSchedule = "@weekly"
)

//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=csiaddons.openshift.io,resources=reclaimspacecronjobs,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *AnnotationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("starting")
	// Fetch PersistentVolumeClaim instance
	pvc := &corev1.PersistentVolumeClaim{}
	err := r.Client.Get(ctx, req.NamespacedName, pvc)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			logger.Info("PersistentVolumeClaim resource not found")

			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}
	logger.Info(rsCronJobScheduleTimeAnnotation)
	found, schedule := getScheduleFromAnnotation(&logger, pvc)
	if !found {
		logger.Info("Annotation not set, exiting reconcile")
		//no annotation set, dequeue
		return ctrl.Result{}, nil
	}

	rsCronJobName := fmt.Sprintf("%s-%v", pvc.Name, pvc.UID)

	rsCronJob := &csiaddonsv1alpha1.ReclaimSpaceCronJob{}
	namespacedName := types.NamespacedName{
		Namespace: req.Namespace,
		Name:      rsCronJobName,
	}
	err = r.Client.Get(ctx, namespacedName, rsCronJob)
	if err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, "Failed to get reclaimSpaceCronJob")

		return ctrl.Result{}, err
	}

	rsCronJob.Spec = csiaddonsv1alpha1.ReclaimSpaceCronJobSpec{
		Schedule: schedule,
		JobSpec: csiaddonsv1alpha1.ReclaimSpaceJobTemplateSpec{
			Spec: csiaddonsv1alpha1.ReclaimSpaceJobSpec{
				Target:               csiaddonsv1alpha1.TargetSpec{PersistentVolumeClaim: pvc.Name},
				BackoffLimit:         10,
				RetryDeadlineSeconds: 800,
			},
		},
	}

	if !apierrors.IsNotFound(err) {
		err = r.Client.Update(ctx, rsCronJob)
		if err != nil {
			logger.Error(err, "Failed to update reclaimSpaceCronJob")

			return ctrl.Result{}, err
		}

		logger.Info("Successfully updated reclaimSpaceCronJob")
		return ctrl.Result{}, nil
	}

	rsCronJob.Name = rsCronJobName
	rsCronJob.Namespace = req.Namespace
	err = ctrl.SetControllerReference(pvc, rsCronJob, r.Scheme)
	if err != nil {
		logger.Error(err, "Failed to set ControllerReference")

		return ctrl.Result{}, err
	}

	err = r.Client.Create(ctx, rsCronJob)
	if err != nil {
		logger.Error(err, "Failed to create reclaimSpaceCronJob")

		return ctrl.Result{}, err
	}

	annontations := pvc.GetAnnotations()
	annontations[rsCronJobNameAnnotation] = rsCronJobName
	pvc.SetAnnotations(annontations)
	err = r.Client.Update(ctx, pvc)
	if err != nil {
		logger.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	logger.Info("Successfully created reclaimSpaceCronJob")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AnnotationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.PersistentVolumeClaim{}).
		Owns(&csiaddonsv1alpha1.ReclaimSpaceCronJob{}).
		WithEventFilter(predicate.AnnotationChangedPredicate{}).
		Complete(r)
}

func getScheduleFromAnnotation(logger *logr.Logger, obj client.Object) (bool, string) {
	schedule, ok := obj.GetAnnotations()[rsCronJobScheduleTimeAnnotation]
	if !ok {
		return false, ""
	}
	_, err := cron.ParseStandard(schedule)
	if err != nil {
		logger.Info(fmt.Sprintf("Parsing given schedule %q failed, using default schedule %q",
			schedule,
			defaultSchedule),
			"error",
			err)

		return true, defaultSchedule
	}
	return true, schedule
}
