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

	csiaddonsv1alpha1 "github.com/csi-addons/kubernetes-csi-addons/api/v1alpha1"
	conn "github.com/csi-addons/kubernetes-csi-addons/internal/connection"
	"github.com/csi-addons/kubernetes-csi-addons/internal/proto"
	"github.com/csi-addons/spec/lib/go/identity"
	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	scv1 "k8s.io/api/storage/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ReclaimSpaceJobReconciler reconciles a ReclaimSpaceJob object
type ReclaimSpaceJobReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	ConnPool *conn.ConnectionPool
}

//+kubebuilder:rbac:groups=csiaddons.openshift.io,resources=reclaimspacejobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=csiaddons.openshift.io,resources=reclaimspacejobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=csiaddons.openshift.io,resources=reclaimspacejobs/finalizers,verbs=update
//+kubebuilder:rbac:groups=v1,resources=persistentvolumeclaim,verbs=get
//+kubebuilder:rbac:groups=v1,resources=persistentvolume,verbs=get
//+kubebuilder:rbac:groups=storage.k8s.io/v1,resources=volumeattachments,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ReclaimSpaceJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *ReclaimSpaceJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch VolumeReplication instance
	instance := &csiaddonsv1alpha1.ReclaimSpaceJob{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			logger.Info("ReclaimSpaceJob resource not found")

			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
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

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReclaimSpaceJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&csiaddonsv1alpha1.ReclaimSpaceJob{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
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
		status.Result = csiaddonsv1alpha1.OperationResultFailed
		status.Message = "retryLimitReached"
		//setFailedCondition
		//setStatus(status,Result,Message,ReclaimSpace)
		return nil
	}

	if time.Now().After(status.StartTime.Time.Add(time.Second * time.Duration(spec.RetryDeadlineSeconds))) {
		status.Result = csiaddonsv1alpha1.OperationResultFailed
		status.Message = "timeLimitReached"
		return nil
	}

	logger.Info("here is the status", "Status", status)

	driverName, pvName, err := r.getTargetVolumeName(ctx, &logger, spec.Target, namespace)
	if err != nil {
		logger.Error(err, "failed to get Persistent Volume Name", "PVCName", spec.Target.PersistentVolumeClaim)

		return err
	}

	nodeID, err := r.getNodeName(ctx, &logger, pvName)
	if err != nil {
		logger.Error(err, "failed to get Persistent Volume Name", "PVCName", spec.Target.PersistentVolumeClaim)

		return err
	}

	conns, unlock := r.ConnPool.Get()
	var (
		controllerClient *grpc.ClientConn = nil
		nodeClient       *grpc.ClientConn = nil
	)
	logger.Info("", "o", *conns)
	for k, v := range *conns {
		if v.DriverName == driverName {
			for _, cap := range v.Capabilities {
				if cap.GetReclaimSpace() == nil {
					continue
				}
				client := *v.Client
				logger.Info(cap.String() + " node:" + v.NodeID)
				if cap.GetReclaimSpace().GetType() == identity.Capability_ReclaimSpace_OFFLINE {
					controllerClient = &client
					logger.Info(k + " is a controller")
					logger.Info("making controller req")
					res, err := proto.NewReclaimSpaceClient(controllerClient).ControllerReclaimSpace(ctx, &proto.ReclaimSpaceRequest{
						PvName: pvName,
					})
					if err != nil {
						return err
					}
					logger.Info("res", "res", res)

				} else if (v.NodeID == nodeID) && (cap.GetReclaimSpace().GetType() == identity.Capability_ReclaimSpace_ONLINE) {
					nodeClient = &client
					logger.Info(k + " is a node")
					logger.Info("making node req")
					res, err := proto.NewReclaimSpaceClient(nodeClient).NodeReclaimSpace(ctx, &proto.ReclaimSpaceRequest{
						PvName: pvName,
					})
					if err != nil {
						return err
					}
					logger.Info(res.String())
				}
			}
		}
	}
	unlock()

	// logger.Info("res", "res", res)
	// send csi driver rpc
	return nil
}

func (r *ReclaimSpaceJobReconciler) getTargetVolumeName(ctx context.Context, logger *logr.Logger, target csiaddonsv1alpha1.TargetSpec, namespace string) (string, string, error) {

	req := types.NamespacedName{Name: target.PersistentVolumeClaim, Namespace: namespace}

	pvc := &corev1.PersistentVolumeClaim{}

	err := r.Client.Get(ctx, req, pvc)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Error(err, "PVC not found", "PVC Name", req.Name)
		}
		return "", "", err
	}

	// Validate PVC in bound state
	if pvc.Status.Phase != corev1.ClaimBound {
		return "", "", fmt.Errorf("PVC %q is not bound to any PV", req.Name)
	}

	return pvc.Annotations["volume.beta.kubernetes.io/storage-provisioner"], pvc.Spec.VolumeName, nil
}

func (r *ReclaimSpaceJobReconciler) getNodeName(ctx context.Context, logger *logr.Logger, pvName string) (string, error) {

	req := types.NamespacedName{Name: pvName}
	volumeAttachments := &scv1.VolumeAttachmentList{}
	err := r.Client.List(ctx, volumeAttachments)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Error(err, "PVC not found", "PVC Name", req.Name)
		}
		return "", err
	}

	for _, v := range volumeAttachments.Items {
		if *v.Spec.Source.PersistentVolumeName == pvName {
			return v.Spec.NodeName, nil
		}
	}

	return "", nil
}
