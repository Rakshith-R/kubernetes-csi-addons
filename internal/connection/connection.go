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

	"github.com/csi-addons/spec/lib/go/identity"
	"google.golang.org/grpc"
)

type Connection struct {
	Client       *grpc.ClientConn
	Capabilities []*identity.Capability
	NodeID       string
	PodUID       string
	DriverName   string
}

func NewConnection(endpoint string, nodeID, driverName string) (*Connection, error) {
	opts := grpc.WithInsecure()
	cc, err := grpc.Dial(endpoint, opts)
	if err != nil {
		return nil, err
	}

	identityClient := identity.NewIdentityClient(cc)
	res, err := identityClient.GetCapabilities(context.TODO(), &identity.GetCapabilitiesRequest{})
	if err != nil {
		return nil, err
	}

	return &Connection{
		Client:       cc,
		Capabilities: res.GetCapabilities(),
		NodeID:       nodeID,
		DriverName:   driverName,
	}, nil
}

func (c *Connection) Close() {
	c.Client.Close()
}
