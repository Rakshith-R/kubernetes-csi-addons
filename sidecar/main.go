/*
Copyright 2021.

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
	"time"

	"github.com/csi-addons/kubernetes-csi-addons/sidecar/client"
	"github.com/csi-addons/kubernetes-csi-addons/sidecar/csiaddonsnode"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

func main() {

	fmt.Println("am here 0")
	var (
		defaultTimeout = time.Minute * 3
		// timeout            = flag.Duration("timeout", defaultTimeout, "Timeout for waiting for response")
		csiAddonsAddress   = flag.String("csi-addons-address", "/run/csi-addons/socket", "CSI Addons endopoint")
		nodeID             = flag.String("node-id", "", "NodeID")
		controllerEndpoint = flag.String("controller-endpoint", "", "The TCP network port where the HTTP server for controller request, will listen (example: `8080`)")
	)
	flag.Parse()

	fmt.Println("am here 1")
	client, err := client.New(*csiAddonsAddress, defaultTimeout)
	if err != nil {
		klog.Fatalf("Failed to connect to %q : %v", *csiAddonsAddress, err)
	}

	err = client.Probe()
	if err != nil {
		klog.Fatalf("Failed to probe driver: %v", err)
	}

	driverName, err := client.GetDriverName()
	if err != nil {
		klog.Fatalf("Failed to get driver name: %v", err)
	}

	cfg, err := rest.InClusterConfig()
	if err != nil {
		klog.Fatalf("Failed to get cluster config: %w", err)
	}

	err = csiaddonsnode.Deploy(cfg, driverName, *nodeID, *controllerEndpoint)
	if err != nil {
		klog.Fatalf("Failed to create csiaddonsnode: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Failed to create client: %v", err)
	}
	fmt.Println(clientset)
	time.Sleep(900 * time.Minute)
}
