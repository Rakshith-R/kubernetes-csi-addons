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

package services

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/csi-addons/kubernetes-csi-addons/internal/service/reclaimspace"
	csiReclaimSpace "github.com/csi-addons/spec/lib/go/reclaimspace"
	"google.golang.org/grpc"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// IdentityServer struct of sidecar with supported methods of CSI
// identity server spec.
type ReclaimSpaceServer struct {
	reclaimspace.UnimplementedReclaimSpaceServer
	controllerClient csiReclaimSpace.ReclaimSpaceControllerClient
	nodeClient       csiReclaimSpace.ReclaimSpaceNodeClient
	kubeClient       *kubernetes.Clientset
}

// NewIdentityServer creates a new IdentityServer which handles the Identity
// Service requests from the CSI-Addons specification.
func NewReclaimSpaceControllerServer(c *grpc.ClientConn, kc *kubernetes.Clientset) *ReclaimSpaceServer {
	return &ReclaimSpaceServer{
		controllerClient: csiReclaimSpace.NewReclaimSpaceControllerClient(c),
		nodeClient:       csiReclaimSpace.NewReclaimSpaceNodeClient(c),
		kubeClient:       kc,
	}
}

func (rs *ReclaimSpaceServer) RegisterService(server grpc.ServiceRegistrar) {
	reclaimspace.RegisterReclaimSpaceServer(server, rs)
}

func (rs *ReclaimSpaceServer) ControllerReclaimSpace(
	ctx context.Context,
	req *reclaimspace.ControllerReclaimSpaceRequest) (*reclaimspace.ReclaimSpaceResponse, error) {

	pvName := req.GetPvName()
	klog.Info(pvName)

	//get Following params from the pvName

	csiReq := &csiReclaimSpace.ControllerReclaimSpaceRequest{
		VolumeId:   "",
		Parameters: map[string]string{},
		Secrets:    map[string]string{},
	}
	res, err := rs.controllerClient.ControllerReclaimSpace(ctx, csiReq)
	klog.Info(err)
	return &reclaimspace.ReclaimSpaceResponse{
		PreUsage:  &reclaimspace.StorageConsumption{UsageBytes: res.PreUsage.UsageBytes},
		PostUsage: &reclaimspace.StorageConsumption{UsageBytes: res.PostUsage.UsageBytes},
	}, nil
}

func (rs *ReclaimSpaceServer) NodeReclaimSpace(
	ctx context.Context,
	req *reclaimspace.NodeReclaimSpaceRequest) (*reclaimspace.ReclaimSpaceResponse, error) {

	pvName := req.GetPvName()
	klog.Info(pvName)

	//get Following params from the pvName

	csiReq := &csiReclaimSpace.NodeReclaimSpaceRequest{
		VolumeId:          "",
		VolumePath:        "",
		StagingTargetPath: "",
		VolumeCapability:  &csi.VolumeCapability{},
		Secrets:           map[string]string{},
	}
	res, err := rs.nodeClient.NodeReclaimSpace(ctx, csiReq)
	klog.Info(err)
	return &reclaimspace.ReclaimSpaceResponse{
		PreUsage:  &reclaimspace.StorageConsumption{UsageBytes: res.PreUsage.UsageBytes},
		PostUsage: &reclaimspace.StorageConsumption{UsageBytes: res.PostUsage.UsageBytes},
	}, nil
}
