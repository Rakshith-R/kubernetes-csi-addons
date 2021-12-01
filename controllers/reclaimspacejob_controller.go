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
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"

	scv1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	csiaddonsv1alpha1 "github.com/csi-addons/kubernetes-csi-addons/api/v1alpha1"
)

const (
	reclaimSpaceJob = "ReclaimSpaceJob"
)

// ReclaimSpaceJobReconciler reconciles a ReclaimSpaceJob object
type ReclaimSpaceJobReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=csiaddons.openshift.io,resources=reclaimspacejobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=csiaddons.openshift.io,resources=reclaimspacejobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=csiaddons.openshift.io,resources=reclaimspacejobs/finalizers,verbs=update
//+kubebuilder:rbac:groups=csiaddons.openshift.io,resources=persistentvolumeclaim,verbs=get
//+kubebuilder:rbac:groups=csiaddons.openshift.io,resources=persistentvolume,verbs=get
//+kubebuilder:rbac:groups=csiaddons.openshift.io,resources=volumeattachments,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ReclaimSpaceJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("Request.Name", req.Name, "Request.Namespace", req.Namespace)

	instance := &csiaddonsv1alpha1.ReclaimSpaceJob{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("reclaimSpaceJob resource not found")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	err = r.reconcile(
		ctx,
		&instance.Spec,
		&instance.Status,
		req.Namespace,
		logger)

	instance.Status.Retries++

	if uErr := r.Client.Status().Update(ctx, instance); uErr != nil {
		logger.Error(err, "failed to update status")
		return ctrl.Result{}, uErr
	}

	return ctrl.Result{}, err
}

func (r *ReclaimSpaceJobReconciler) reconcile(
	ctx context.Context,
	spec *csiaddonsv1alpha1.ReclaimSpaceJobSpec,
	status *csiaddonsv1alpha1.ReclaimSpaceJobStatus,
	namespace string,
	logger logr.Logger) error {

	if status.StartTime == nil {
		status.StartTime = &v1.Time{
			Time: time.Now(),
		}
	}

	if status.Result != "" {
		// since result is already set, just dequeue
		return nil
	}

	if status.Retries > spec.BackoffLimit {
		logger.Info("maximum retry limit reached")
		status.Result = csiaddonsv1alpha1.ResultFailed
		status.Message = "retryLimitReached"
		//setFailedCondition
		//setStatus(status,Result,Message,ReclaimSpace)
		return nil
	}

	if time.Now().After(status.StartTime.Time.Add(time.Second * time.Duration(spec.ActiveDeadlineSeconds))) {
		status.Result = csiaddonsv1alpha1.ResultFailed
		status.Message = "timeLimitReached"
		return nil
	}

	logger.Info("here is the status", "Status", status)

	sc, pvc, pv, err := getTargetVolume(ctx, r.Client, r.Log, spec.Target, namespace)
	if err != nil {
		logger.Error(err, "failed to get PVC", "PVCName", spec.Target.PVC)
		return err
	}

	logger.Info("pvc", pvc.Name, pvc.Namespace)
	logger.Info("pv", pv.Name, pv.Kind)
	logger.Info("sc", sc.Name, sc.Kind)
	return nil
}

func getTargetVolume(ctx context.Context, client client.Client, logger logr.Logger, target csiaddonsv1alpha1.TargetSpec, namespace string) (*scv1.StorageClass, *corev1.PersistentVolumeClaim, *corev1.PersistentVolume, error) {

	req := types.NamespacedName{Name: target.PVC, Namespace: namespace}
	pvc := &corev1.PersistentVolumeClaim{}
	err := client.Get(ctx, req, pvc)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Error(err, "PVC not found", "PVC Name", req.Name)
		}
		return nil, nil, nil, err
	}

	// Validate PVC in bound state
	if pvc.Status.Phase != corev1.ClaimBound {
		return nil, nil, nil, fmt.Errorf("PVC %q is not bound to any PV", req.Name)
	}

	// Get PV object for the PVC
	pvName := pvc.Spec.VolumeName
	pv := &corev1.PersistentVolume{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: pvName}, pv)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Error(err, "PV not found", "PV Name", pvName)
		}
		return nil, nil, nil, err
	}

	scName := *pvc.Spec.StorageClassName
	sc := &scv1.StorageClass{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: scName}, sc)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Error(err, "Storageclass not found", "Storageclass Name", scName)
		}
		return nil, nil, nil, err
	}

	return sc, pvc, pv, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReclaimSpaceJobReconciler) SetupWithManager(mgr ctrl.Manager) error {

	err := r.waitForVolumeReplicationResource(reclaimSpaceJob)
	if err != nil {
		r.Log.Error(err, "failed to wait for crds")
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&csiaddonsv1alpha1.ReclaimSpaceJob{}).
		WithEventFilter(predicate.GenerationChangedPredicate{
			Funcs: predicate.Funcs{},
		}).
		Complete(r)
}

func (r *ReclaimSpaceJobReconciler) waitForVolumeReplicationResource(resourceName string) error {
	logger := r.Log.WithName("checkingDependencies")
	unstructuredResource := &unstructured.Unstructured{}

	unstructuredResource.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   csiaddonsv1alpha1.GroupVersion.Group,
		Kind:    resourceName,
		Version: csiaddonsv1alpha1.GroupVersion.Version,
	})
	for {
		err := r.Client.List(context.TODO(), unstructuredResource)
		if err == nil {
			return nil
		}
		// return errors other than NoMatch
		if !meta.IsNoMatchError(err) {
			logger.Error(err, "got an unexpected error while waiting for resource", "Resource", resourceName)
			return err
		}
		logger.Info("resource does not exist", "Resource", resourceName)
		time.Sleep(5 * time.Second)
	}
}
