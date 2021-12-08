module github.com/csi-addons/kubernetes-csi-addons

go 1.16

require (
	github.com/csi-addons/spec v0.1.2-0.20211123125058-fd968c478af7
	github.com/go-logr/logr v0.4.0
	github.com/kubernetes-csi/csi-lib-utils v0.10.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.14.0
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.26.0
	k8s.io/api v0.22.0
	k8s.io/apimachinery v0.22.0
	k8s.io/client-go v0.22.0
	sigs.k8s.io/controller-runtime v0.9.6
)
