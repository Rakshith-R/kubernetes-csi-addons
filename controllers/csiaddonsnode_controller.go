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

	csiaddonsv1alpha1 "github.com/csi-addons/kubernetes-csi-addons/api/v1alpha1"
	conn "github.com/csi-addons/kubernetes-csi-addons/internal/connection"
	"github.com/csi-addons/kubernetes-csi-addons/internal/util"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	csiAddonsNodeFinalizer = "csiaddons.openshift.io"
)

// CSIAddonsNodeReconciler reconciles a CSIAddonsNode object
type CSIAddonsNodeReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	ConnPool *conn.ConnectionPool
}

//+kubebuilder:rbac:groups=csiaddons.openshift.io,resources=csiaddonsnodes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=csiaddons.openshift.io,resources=csiaddonsnodes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=csiaddons.openshift.io,resources=csiaddonsnodes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CSIAddonsNode object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *CSIAddonsNodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch VolumeReplication instance
	instance := &csiaddonsv1alpha1.CSIAddonsNode{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			logger.Info("CSIAddonsNode resource not found")

			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	nodeID := instance.Spec.Driver.NodeID
	driverName := instance.Spec.Driver.Name
	endPoint := instance.Spec.Driver.EndPoint
	key := r.creatKey(instance.Namespace, instance.Name)

	logger = logger.WithValues(
		"NodeID", nodeID,
		"EndPoint", endPoint,
		"Driver Name", driverName,
		"key", key)

	if !instance.DeletionTimestamp.IsZero() {
		logger.Info("deleting connection")

		r.ConnPool.Delete(key)

		if util.ContainsInSlice(instance.Finalizers, csiAddonsNodeFinalizer) {
			logger.Info("removing finalizer")

			instance.Finalizers = util.RemoveFromSlice(instance.Finalizers, csiAddonsNodeFinalizer)
			if err = r.Client.Update(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, nil
		}
	}

	if !util.ContainsInSlice(instance.Finalizers, csiAddonsNodeFinalizer) {
		logger.Info("adding finalizer")
		instance.Finalizers = append(instance.Finalizers, csiAddonsNodeFinalizer)
		if err = r.Client.Update(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	logger.Info("connecting to sidecar")
	newCon, err := conn.NewConnection(ctx,
		endPoint,
		nodeID,
		driverName)
	if err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("successfully connected to sidecar")
	r.ConnPool.Put(r.creatKey(req.Namespace, req.Name), newCon)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CSIAddonsNodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&csiaddonsv1alpha1.CSIAddonsNode{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}

func (r *CSIAddonsNodeReconciler) creatKey(namespace, name string) string {
	return namespace + "/" + name
}
