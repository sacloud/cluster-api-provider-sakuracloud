module github.com/sacloud/cluster-api-provider-sakuracloud

go 1.12

require (
	github.com/go-logr/logr v0.1.0
	github.com/google/uuid v1.1.1 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/pkg/errors v0.8.1
	github.com/sacloud/ftps v0.0.0-20171205062625-42fc0f9886fe
	github.com/sacloud/libsacloud/v2 v2.0.0-beta5.0.20191011051923-d3fd15b18992
	k8s.io/api v0.0.0-20190918195907-bd6ac527cfd2
	k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/cluster-bootstrap v0.0.0-20190711112844-b7409fb13d1b
	k8s.io/klog v0.4.0
	sigs.k8s.io/cluster-api v0.2.3
	sigs.k8s.io/cluster-api-bootstrap-provider-kubeadm v0.1.1
	sigs.k8s.io/controller-runtime v0.2.2
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190704095032-f4ca3d3bdf1d
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190918200256-06eb1244587a
	sigs.k8s.io/cluster-api => sigs.k8s.io/cluster-api v0.2.0
	sigs.k8s.io/cluster-api-bootstrap-provider-kubeadm => sigs.k8s.io/cluster-api-bootstrap-provider-kubeadm v0.1.1
)
