// Copyright 2019 The Operator-SDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configmap

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/operator-framework/operator-sdk/internal/util/k8sutil"
)

const (
	// The image operator-registry's initializer and registry-server binaries are run from.
	// This iamge has been pinned to the latest version containing the initializer and registry-server
	// binaries, which were deprecated in https://github.com/operator-framework/operator-registry/pull/587.
	registryBaseImage = "quay.io/operator-framework/upstream-registry-builder:v1.16.0"
	// The port registry-server will listen on within a container.
	registryGRPCPort = 50051
	// Path of the bundle database generated by initializer. Use /tmp since it is
	// typically world-writable.
	registryDBName = "/tmp/bundle.db"
	// Path of the log file generated by registry-server. Use /tmp since it is
	// typically world-writable.
	registryLogFile = "/tmp/termination.log"
)

func getRegistryServerName(pkgName string) string {
	name := k8sutil.FormatOperatorNameDNS1123(pkgName)
	return fmt.Sprintf("%s-registry-server", name)
}

// getRegistryDeploymentLabels creates a set of labels to identify
// operator-registry Deployment objects.
func getRegistryDeploymentLabels(pkgName string) map[string]string {
	labels := makeRegistryLabels(pkgName)
	labels["server-name"] = getRegistryServerName(pkgName)
	return labels
}

// applyToDeploymentPodSpec applies f to dep's pod template spec.
func applyToDeploymentPodSpec(dep *appsv1.Deployment, f func(*corev1.PodSpec)) {
	f(&dep.Spec.Template.Spec)
}

// withConfigMapVolume returns a function that appends a volume with name
// volName containing a reference to a ConfigMap with name cmName to the
// Deployment argument's pod template spec.
func withConfigMapVolume(volName, cmName string) func(*appsv1.Deployment) {
	volume := corev1.Volume{
		Name: volName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cmName,
				},
			},
		},
	}
	return func(dep *appsv1.Deployment) {
		applyToDeploymentPodSpec(dep, func(spec *corev1.PodSpec) {
			spec.Volumes = append(spec.Volumes, volume)
		})
	}
}

// withContainerVolumeMounts returns a function that appends volumeMounts
// to each container in the Deployment argument's pod template spec. One
// volumeMount is appended for each path in paths from volume with name
// volName.
func withContainerVolumeMounts(volName string, paths ...string) func(*appsv1.Deployment) {
	volumeMounts := []corev1.VolumeMount{}
	for _, p := range paths {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      volName,
			MountPath: p,
		})
	}
	return func(dep *appsv1.Deployment) {
		applyToDeploymentPodSpec(dep, func(spec *corev1.PodSpec) {
			for i := range spec.Containers {
				spec.Containers[i].VolumeMounts = append(spec.Containers[i].VolumeMounts, volumeMounts...)
			}
		})
	}
}

// getDBContainerCmd returns a command string that, when run, does two things:
//  1. Runs a database initializer on the manifests in the /registry
//     directory.
//  2. Runs an operator-registry server serving the bundle database.
func getDBContainerCmd(dbPath, logPath string) string {
	initCmd := fmt.Sprintf("/bin/initializer -o %s -m %s", dbPath, containerManifestsDir)
	srvCmd := fmt.Sprintf("/bin/registry-server -d %s -t %s", dbPath, logPath)
	return fmt.Sprintf("%s && %s", initCmd, srvCmd)
}

// withRegistryGRPCContainer returns a function that appends a container
// running an operator-registry GRPC server to the Deployment argument's
// pod template spec.
func withRegistryGRPCContainer(pkgName string) func(*appsv1.Deployment) {
	container := corev1.Container{
		Name:       getRegistryServerName(pkgName),
		Image:      registryBaseImage,
		WorkingDir: "/tmp",
		Command:    []string{"/bin/sh"},
		Args: []string{
			"-c",
			// TODO(estroz): grab logs and print if error
			getDBContainerCmd(registryDBName, registryLogFile),
		},
		Ports: []corev1.ContainerPort{
			{Name: "registry-grpc", ContainerPort: registryGRPCPort},
		},
	}
	return func(dep *appsv1.Deployment) {
		applyToDeploymentPodSpec(dep, func(spec *corev1.PodSpec) {
			spec.Containers = append(spec.Containers, container)
		})
	}
}

// newRegistryDeployment creates a new Deployment with a name derived from
// pkgName, the package manifest's packageName, in namespace. The Deployment
// and replicas are created with labels derived from pkgName. opts will be
// applied to the Deployment object.
func newRegistryDeployment(pkgName, namespace string, opts ...func(*appsv1.Deployment)) *appsv1.Deployment {
	var replicas int32 = 1
	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      getRegistryServerName(pkgName),
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: getRegistryDeploymentLabels(pkgName),
			},
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: getRegistryDeploymentLabels(pkgName),
				},
			},
		},
	}
	for _, opt := range opts {
		opt(dep)
	}
	return dep
}
