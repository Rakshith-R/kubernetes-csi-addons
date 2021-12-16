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

package connection

import (
	"context"
	"time"

	"github.com/csi-addons/spec/lib/go/identity"
	"google.golang.org/grpc"
)

const defaultTimeout = time.Minute

type Connection struct {
	Client       *grpc.ClientConn
	Capabilities []*identity.Capability
	NodeID       string
	DriverName   string
}

func NewConnection(ctx context.Context, endpoint string, nodeID, driverName string) (*Connection, error) {
	opts := grpc.WithInsecure()
	cc, err := grpc.Dial(endpoint, opts)
	if err != nil {
		return nil, err
	}

	caps, err := getCapabilities(ctx, cc)
	if err != nil {
		return nil, err
	}
	return &Connection{
		Client:       cc,
		Capabilities: caps,
		NodeID:       nodeID,
		DriverName:   driverName,
	}, nil
}

func (c *Connection) Close() {
	c.Client.Close()
}

func getCapabilities(ctx context.Context, cc *grpc.ClientConn) ([]*identity.Capability, error) {
	newCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	identityClient := identity.NewIdentityClient(cc)
	res, err := identityClient.GetCapabilities(newCtx, &identity.GetCapabilitiesRequest{})
	if err != nil {
		return nil, err
	}
	return res.GetCapabilities(), nil
}
