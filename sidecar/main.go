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

package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/csi-addons/kubernetes-csi-addons/sidecar/client"
	"github.com/csi-addons/kubernetes-csi-addons/sidecar/server"
)

func main() {
	var (
		defaultTimeout = time.Minute * 3
		timeout        = flag.Duration("rpc-timeout", defaultTimeout, "timeout")
		// kubeconfig  = flag.String("kubeconfig", "~/.kube/config", "Absolute path to the kubeconfig file")
		// Endpoint    = flag.String("http-endpoint", "", "The TCP network address where the HTTP server")
		// DriverName  = flag.String("drivername", "", "The driver name")
		nodeID           = flag.String("node-id", "", "The node-id on which the pod is running")
		csiAddonsAddress = flag.String("csi-addons-address", "/csi/csi-provisioner.sock", "The gRPC endpoint for Target CSI Volume.")
	)
	flag.Parse()
	fmt.Println(*nodeID)
	fmt.Println(*csiAddonsAddress)
	client, err := client.New(*csiAddonsAddress, *timeout)
	if err != nil {
		log.Fatalf("connecting failed", err)
	}
	drivername, err := client.GetDriverName()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(drivername)

	server.Serve()
}
