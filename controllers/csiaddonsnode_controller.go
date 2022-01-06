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

	csiaddonsv1alpha1 "github.com/csi-addons/kubernetes-csi-addons/api/v1alpha1"
	"github.com/csi-addons/kubernetes-csi-addons/internal/connection"
	"github.com/csi-addons/kubernetes-csi-addons/internal/util"
	"github.com/go-logr/logr"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	csiAddonsNodeFinalizer = "csiaddons.openshift.io/csiaddonsnode"
)

// CSIAddonsNodeReconciler reconciles a CSIAddonsNode object
type CSIAddonsNodeReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	ConnPool *connection.ConnectionPool
}

//+kubebuilder:rbac:groups=csiaddons.openshift.io,resources=csiaddonsnodes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=csiaddons.openshift.io,resources=csiaddonsnodes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=csiaddons.openshift.io,resources=csiaddonsnodes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *CSIAddonsNodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch CSIAddonsNode instance
	instance := &csiaddonsv1alpha1.CSIAddonsNode{}
	if err := r.Client.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			logger.Info("CSIAddonsNode resource not found")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	nodeID := instance.Spec.Driver.NodeID
	driverName := instance.Spec.Driver.Name
	endPoint := instance.Spec.Driver.EndPoint
	key := r.createKey(instance.Namespace, instance.Name)
	logger = logger.WithValues("NodeID", nodeID, "EndPoint", endPoint, "Driver Name", driverName)

	if !instance.DeletionTimestamp.IsZero() {
		logger.Info("Deleting connection")
		r.ConnPool.Delete(key)
		if err := r.removeFinalizer(ctx, instance, &logger); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if err := r.addFinalizer(ctx, instance, &logger); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("Connecting to sidecar")
	newCon, err := connection.NewConnection(ctx, endPoint, nodeID, driverName)
	if err != nil {
		logger.Error(err, "Failed to establish connection with sidecar")

		errMessage := util.GetErrorMessage(err)
		instance.Status.State = csiaddonsv1alpha1.CSIAddonsNodeStateFailed
		instance.Status.Message = fmt.Sprintf("Failed to establish connection with sidecar: %v", errMessage)
		if statusErr := r.Client.Status().Update(ctx, instance); statusErr != nil {
			logger.Error(statusErr, "Failed to update status")
			return ctrl.Result{}, statusErr
		}

		return ctrl.Result{}, err
	}

	logger.Info("Successfully connected to sidecar")
	r.ConnPool.Put(key, newCon)
	logger.Info("Added connection to connection pool")

	instance.Status.State = csiaddonsv1alpha1.CSIAddonsNodeStateConnected
	instance.Status.Message = "Successfully established connection with sidecar"
	if err = r.Client.Status().Update(ctx, instance); err != nil {
		logger.Error(err, "Failed to update status")

		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CSIAddonsNodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&csiaddonsv1alpha1.CSIAddonsNode{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}

// addFinalizer adds finalizer to object instance if it is not present.
func (r *CSIAddonsNodeReconciler) addFinalizer(
	ctx context.Context,
	instance *csiaddonsv1alpha1.CSIAddonsNode,
	logger *logr.Logger) error {

	if !util.ContainsInSlice(instance.Finalizers, csiAddonsNodeFinalizer) {
		logger.Info("Adding finalizer")

		instance.Finalizers = append(instance.Finalizers, csiAddonsNodeFinalizer)
		if err := r.Client.Update(ctx, instance); err != nil {
			logger.Error(err, "Failed to add finalizer")

			return err
		}
	}

	return nil
}

// removeFinalizer removes finalizer to object instance if it is present.
func (r *CSIAddonsNodeReconciler) removeFinalizer(
	ctx context.Context,
	instance *csiaddonsv1alpha1.CSIAddonsNode,
	logger *logr.Logger) error {

	if util.ContainsInSlice(instance.Finalizers, csiAddonsNodeFinalizer) {
		logger.Info("Removing finalizer")

		instance.Finalizers = util.RemoveFromSlice(instance.Finalizers, csiAddonsNodeFinalizer)
		if err := r.Client.Update(ctx, instance); err != nil {
			logger.Error(err, "Failed to remove finalizer")

			return err
		}
	}
	return nil
}

// createKey returns "namespace/name" string to be used as key for connection pool.
func (r *CSIAddonsNodeReconciler) createKey(namespace, name string) string {
	return namespace + "/" + name
}
