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
	"errors"
	"fmt"
	"math"
	"time"

	csiaddonsv1alpha1 "github.com/csi-addons/kubernetes-csi-addons/api/v1alpha1"
	"github.com/csi-addons/kubernetes-csi-addons/internal/connection"
	"github.com/csi-addons/kubernetes-csi-addons/internal/proto"
	"github.com/csi-addons/spec/lib/go/identity"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	scv1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
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
	ConnPool *connection.ConnectionPool
	Timeout  time.Duration
}

//+kubebuilder:rbac:groups=csiaddons.openshift.io,resources=reclaimspacejobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=csiaddons.openshift.io,resources=reclaimspacejobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=csiaddons.openshift.io,resources=reclaimspacejobs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get,list,watch
//+kubebuilder:rbac:groups="",resources=persistentvolumes,verbs=get,list,watch
//+kubebuilder:rbac:groups=storage.k8s.io,resources=volumeattachments,verbs=get,list,watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ReclaimSpaceJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch ReclaimSpaceJob instance
	rsJob := &csiaddonsv1alpha1.ReclaimSpaceJob{}
	err := r.Client.Get(ctx, req.NamespacedName, rsJob)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			logger.Info("ReclaimSpaceJob resource not found")

			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	if !rsJob.DeletionTimestamp.IsZero() {
		logger.Info("ReclaimSpaceJob resource is being deleted, exiting reconcile")
		return ctrl.Result{}, nil
	}

	if rsJob.Status.Result != "" {
		logger.Info(fmt.Sprintf("ReclaimSpaceJob is already in %q state, exiting reconcile",
			rsJob.Status.Result))
		// since result is already set, just dequeue
		return ctrl.Result{}, nil
	}

	err = validateReclaimSpaceJobSpec(rsJob)
	if err != nil {
		logger.Error(err, "failed to validate ReclaimSpaceJob.Spec")

		rsJob.Status.Result = csiaddonsv1alpha1.OperationResultFailed
		rsJob.Status.Message = fmt.Sprintf("failed to validate ReclaimSpaceJob.Spec: %v", err)
		if statusErr := r.Client.Status().Update(ctx, rsJob); statusErr != nil {
			logger.Error(err, "failed to update status")
			return ctrl.Result{}, statusErr
		}

		// invalid parameter, do not requeue.
		return ctrl.Result{}, nil
	}

	err = r.reconcile(
		ctx,
		&logger,
		&rsJob.Spec,
		&rsJob.Status,
		req.Namespace,
	)

	if rsJob.Status.Result == "" && rsJob.Status.Retries == rsJob.Spec.BackoffLimit {
		logger.Info("Maximum retry limit reached")
		rsJob.Status.Result = csiaddonsv1alpha1.OperationResultFailed
		rsJob.Status.Message = "Maximum retry limit reached"
		rsJob.Status.CompletionTime = &v1.Time{Time: time.Now()}
	}

	if statusErr := r.Client.Status().Update(ctx, rsJob); statusErr != nil {
		logger.Error(statusErr, "failed to update status")

		return ctrl.Result{}, statusErr
	}

	if rsJob.Status.Result != "" {
		// since result is already set, just dequeue
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, err
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
	logger *logr.Logger,
	spec *csiaddonsv1alpha1.ReclaimSpaceJobSpec,
	status *csiaddonsv1alpha1.ReclaimSpaceJobStatus,
	namespace string) error {

	if status.StartTime == nil {
		status.StartTime = &v1.Time{Time: time.Now()}
	} else {
		// not first reconcile, increment retries
		status.Retries++
	}

	if time.Now().After(status.StartTime.Time.Add(time.Second * time.Duration(spec.RetryDeadlineSeconds))) {
		status.Result = csiaddonsv1alpha1.OperationResultFailed
		status.Message = "Time limit reached"
		status.CompletionTime = &v1.Time{Time: time.Now()}

		return nil
	}

	driverName, pvName, nodeID, err := r.getTargetDetails(ctx, logger, spec.Target, namespace)
	if err != nil {
		logger.Error(err, "Failed to get target details")
		status.Message = fmt.Sprintf("Failed to get target details: %v", err)
		//setFailedStatus

		return err
	}

	controllerFound, controllerReclaimedSpace, err := r.controllerReclaimSpace(ctx, logger, driverName, pvName)
	if err != nil {
		logger.Error(err, "Failed to make controller request")
		//setFailedStatus

		return err
	}

	nodeFound, nodeReclaimedSpace, err := r.nodeReclaimSpace(ctx, logger, driverName, nodeID, pvName)
	if err != nil {
		logger.Error(err, "Failed to make node request")
		//setFailedStatus

		return err
	}

	if !controllerFound && !nodeFound {
		//setFailedStatus

		return errors.New("Controller and Node Client not found")
	}

	reclaimedSpace := int64(0)
	if controllerFound && controllerReclaimedSpace != nil {
		reclaimedSpace = *controllerReclaimedSpace
	}
	if nodeFound && nodeReclaimedSpace != nil {
		reclaimedSpace += *nodeReclaimedSpace
	}

	status.Result = csiaddonsv1alpha1.OperationResultSucceeded
	status.Message = "Reclaim Space operation successfully completed."
	status.ReclaimedSpace = *resource.NewQuantity(reclaimedSpace, resource.DecimalSI)

	return nil
}

func (r *ReclaimSpaceJobReconciler) getTargetDetails(
	ctx context.Context,
	logger *logr.Logger,
	target csiaddonsv1alpha1.TargetSpec,
	namespace string) (string, string, string, error) {
	*logger = logger.WithValues("PVCName", target.PersistentVolumeClaim, "PVCNamespace", namespace)
	req := types.NamespacedName{Name: target.PersistentVolumeClaim, Namespace: namespace}
	pvc := &corev1.PersistentVolumeClaim{}

	err := r.Client.Get(ctx, req, pvc)
	if err != nil {
		return "", "", "", err
	}

	// Validate PVC in bound state
	if pvc.Status.Phase != corev1.ClaimBound {
		return "", "", "", fmt.Errorf("PVC %q is not bound to any PV", req.Name)
	}

	*logger = logger.WithValues("PVName", pvc.Spec.VolumeName)
	pv := &corev1.PersistentVolume{}
	req = types.NamespacedName{Name: pvc.Spec.VolumeName}

	err = r.Client.Get(ctx, req, pv)
	if err != nil {
		return "", "", "", err
	}

	if pv.Spec.CSI == nil {
		return "", "", "", errors.New("PersistentVolume.Spec.CSI is nil")
	}

	volumeAttachments := &scv1.VolumeAttachmentList{}
	err = r.Client.List(ctx, volumeAttachments)
	if err != nil {
		return "", "", "", err
	}

	for _, v := range volumeAttachments.Items {
		if *v.Spec.Source.PersistentVolumeName == pv.Name {
			*logger = logger.WithValues("NodeID", v.Spec.NodeName)
			return pv.Spec.CSI.Driver, pv.Name, v.Spec.NodeName, nil
		}
	}

	// return "" for nodeID since volume may not be mounted.
	return pv.Spec.CSI.Driver, pv.Name, "", nil
}

func (r *ReclaimSpaceJobReconciler) getRSClientWithCap(
	driverName, nodeID string,
	capType identity.Capability_ReclaimSpace_Type) (string, proto.ReclaimSpaceClient) {
	conns := r.ConnPool.GetByNodeID(driverName, nodeID)
	for k, v := range conns {
		for _, cap := range v.Capabilities {
			if cap.GetReclaimSpace() == nil {
				continue
			}
			if cap.GetReclaimSpace().Type == capType {
				return k, proto.NewReclaimSpaceClient(v.Client)
			}
		}
	}

	return "", nil
}

func (r *ReclaimSpaceJobReconciler) controllerReclaimSpace(
	ctx context.Context,
	logger *logr.Logger,
	driverName, pvName string) (bool, *int64, error) {
	clientName, controllerClient := r.getRSClientWithCap(driverName, "", identity.Capability_ReclaimSpace_OFFLINE)
	if controllerClient == nil {
		logger.Info("Controller Client not found")
		return false, nil, nil
	}
	*logger = logger.WithValues("controllerClient", clientName)

	req := &proto.ReclaimSpaceRequest{
		PvName: pvName,
	}
	newCtx, cancel := context.WithTimeout(ctx, r.Timeout)
	defer cancel()
	resp, err := controllerClient.ControllerReclaimSpace(newCtx, req)
	if err != nil {
		return true, nil, err
	}

	return true, calculateReclaimedSpace(resp.PreUsage, resp.PostUsage), nil
}

func (r *ReclaimSpaceJobReconciler) nodeReclaimSpace(
	ctx context.Context,
	logger *logr.Logger,
	driverName, nodeID, pvName string) (bool, *int64, error) {
	clientName, nodeClient := r.getRSClientWithCap(driverName, nodeID, identity.Capability_ReclaimSpace_ONLINE)
	if nodeClient == nil {
		logger.Info("Node Client not found")
		return false, nil, nil
	}
	*logger = logger.WithValues("nodeClient", clientName)

	req := &proto.ReclaimSpaceRequest{
		PvName: pvName,
	}
	newCtx, cancel := context.WithTimeout(ctx, r.Timeout)
	defer cancel()
	resp, err := nodeClient.NodeReclaimSpace(newCtx, req)
	if err != nil {
		return true, nil, err
	}

	return true, calculateReclaimedSpace(resp.PreUsage, resp.PostUsage), nil
}

func calculateReclaimedSpace(PreUsage, PostUsage *proto.StorageConsumption) *int64 {
	var (
		preUsage  int64 = 0
		postUsage int64 = 0
	)
	if PreUsage != nil {
		preUsage = PreUsage.UsageBytes
	}
	if PostUsage != nil {
		postUsage = PostUsage.UsageBytes
	}

	result := int64(math.Max(float64(postUsage)-float64(preUsage), 0))
	return &result
}

func validateReclaimSpaceJobSpec(
	rsJob *csiaddonsv1alpha1.ReclaimSpaceJob) error {
	if rsJob.Spec.Target.PersistentVolumeClaim == "" {
		return errors.New("required parameter 'PersistentVolumeClaim' in ReclaimSpaceJob.Spec.Target is empty")
	}

	return nil
}
