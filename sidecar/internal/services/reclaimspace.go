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

	"github.com/csi-addons/spec/lib/go/reclaimspace"
	"google.golang.org/grpc"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// IdentityServer struct of sidecar with supported methods of CSI
// identity server spec.
type ReclaimSpaceControllerServer struct {
	reclaimspace.UnimplementedReclaimSpaceControllerServer
	client     reclaimspace.ReclaimSpaceControllerClient
	kubeClient *kubernetes.Clientset
}

// NewIdentityServer creates a new IdentityServer which handles the Identity
// Service requests from the CSI-Addons specification.
func NewReclaimSpaceControllerServer(c *grpc.ClientConn, kc *kubernetes.Clientset) *ReclaimSpaceControllerServer {
	return &ReclaimSpaceControllerServer{
		client:     reclaimspace.NewReclaimSpaceControllerClient(c),
		kubeClient: kc,
	}
}

func (rs *ReclaimSpaceControllerServer) RegisterService(server grpc.ServiceRegistrar) {
	reclaimspace.RegisterReclaimSpaceControllerServer(server, rs)
}

// GetIdentity returns available capabilities of the rbd driver.
func (rs *ReclaimSpaceControllerServer) ControllerReclaimSpace(
	ctx context.Context,
	req *reclaimspace.ControllerReclaimSpaceRequest) (*reclaimspace.ControllerReclaimSpaceResponse, error) {

	res, err := rs.client.ControllerReclaimSpace(ctx, req)
	klog.Info(err)
	return res, nil
}

type ReclaimSpaceNodeServer struct {
	reclaimspace.UnimplementedReclaimSpaceNodeServer
	client     reclaimspace.ReclaimSpaceNodeClient
	kubeClient *kubernetes.Clientset
}

func NewReclaimSpaceNodeServer(c *grpc.ClientConn, kc *kubernetes.Clientset) *ReclaimSpaceNodeServer {
	return &ReclaimSpaceNodeServer{
		client:     reclaimspace.NewReclaimSpaceNodeClient(c),
		kubeClient: kc,
	}
}

func (rs *ReclaimSpaceNodeServer) RegisterService(server grpc.ServiceRegistrar) {
	reclaimspace.RegisterReclaimSpaceNodeServer(server, rs)
}

func (rs *ReclaimSpaceNodeServer) NodeReclaimSpace(
	ctx context.Context,
	req *reclaimspace.NodeReclaimSpaceRequest) (*reclaimspace.NodeReclaimSpaceResponse, error) {

	res, err := rs.client.NodeReclaimSpace(ctx, req)
	klog.Info(err)
	return res, nil
}
