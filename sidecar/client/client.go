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

package client

import (
	"context"
	"fmt"
	"time"

	"github.com/csi-addons/spec/lib/go/identity"
	"github.com/csi-addons/spec/lib/go/reclaimspace"
	"github.com/kubernetes-csi/csi-lib-utils/connection"
	"github.com/kubernetes-csi/csi-lib-utils/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

const (
	// Interval of trying to call Probe() until it succeeds
	probeInterval = 1 * time.Second
)

// Client holds the GRPC connenction details
type Client struct {
	identity.IdentityClient
	reclaimspace.ReclaimSpaceControllerClient
	Client  *grpc.ClientConn
	Timeout time.Duration
}

// Connect to the GRPC client
func connect(address string) (*grpc.ClientConn, error) {
	return connection.Connect(address, metrics.NewCSIMetricsManager(""), connection.OnConnectionLoss(connection.ExitOnConnectionLoss()))
}

// New creates and returns the GRPC client
func New(address string, timeout time.Duration) (*Client, error) {
	c := &Client{}
	cc, err := connect(address)
	if err != nil {
		return c, err
	}
	c.Client = cc
	c.Timeout = timeout
	return c, nil
}

// Probe the GRPC client once
func (c *Client) Probe() error {
	return probeForever(c.Client, c.Timeout)
}

// GetDriverName gets the driver name from the driver
func (c *Client) GetDriverName() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	req := identity.GetIdentityRequest{}
	rsp, err := c.GetIdentity(ctx, &req)
	if err != nil {
		return "", err
	}

	name := rsp.GetName()
	if name == "" {
		return "", fmt.Errorf("driver name is empty")
	}

	return name, nil
}

// PpobeForever calls Probe() of a CSI driver and waits until the driver becomes ready.
// Any error other than timeout is returned.
func probeForever(conn *grpc.ClientConn, singleProbeTimeout time.Duration) error {
	for {
		klog.Info("Probing CSI driver for readiness")
		ready, err := probeOnce(conn, singleProbeTimeout)
		if err != nil {
			st, ok := status.FromError(err)
			if !ok {
				// This is not gRPC error. The probe must have failed before gRPC
				// method was called, otherwise we would get gRPC error.
				return fmt.Errorf("CSI driver probe failed: %s", err)
			}
			if st.Code() != codes.DeadlineExceeded {
				return fmt.Errorf("CSI driver probe failed: %s", err)
			}
			// Timeout -> driver is not ready. Fall through to sleep() below.
			klog.Warning("CSI driver probe timed out")
		} else {
			if ready {
				return nil
			}
			klog.Warning("CSI driver is not ready")
		}
		// Timeout was returned or driver is not ready.
		time.Sleep(probeInterval)
	}
}

// probeOnce is a helper to simplify defer cancel()
func probeOnce(conn *grpc.ClientConn, timeout time.Duration) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return probe(ctx, conn)
}

// Probe calls driver Probe() just once and returns its result without any processing.
func probe(ctx context.Context, conn *grpc.ClientConn) (ready bool, err error) {
	client := identity.NewIdentityClient(conn)

	req := identity.ProbeRequest{}
	rsp, err := client.Probe(ctx, &req)

	if err != nil {
		return false, err
	}

	r := rsp.GetReady()
	if r == nil {
		// "If not present, the caller SHALL assume that the plugin is in a ready state"
		return true, nil
	}
	return r.GetValue(), nil
}
