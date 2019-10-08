/*
Copyright 2019 Kazumichi Yamamoto.

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

package constants

import (
	"github.com/sacloud/cluster-api-provider-sakuracloud/api/v1alpha2"
)

const (
	// CloudProviderSecretName is the name of the Secret that stores the
	// cloud provider credentials.
	CloudProviderSecretName = "cloud-provider-sakuracloud-credentials"

	// CloudProviderSecretNamespace is the namespace in which the cloud provider
	// credentials secret is located.
	CloudProviderSecretNamespace = "kube-system"

	// DefaultBindPort is the default API port used to generate the kubeadm
	// configurations.
	DefaultBindPort = 6443

	// SakuraCloudCredentialSecretTokenKey is the key used to store/retrieve the
	// SakuraCloud API token from a Kubernetes secret.
	SakuraCloudCredentialSecretTokenKey = "accessToken"

	// SakuraCloudCredentialSecretSecretKey is the key used to store/retrieve the
	// SakuraCloud API secret from a Kubernetes secret.
	SakuraCloudCredentialSecretSecretKey = "accessSecret"

	// MachineReadyAnnotationLabel is the annotation used to indicate that a
	// machine is ready.
	MachineReadyAnnotationLabel = "caps." + v1alpha2.GroupName + "/machine-ready"

	// MaintenanceAnnotationLabel is the annotation used to indicate a machine and/or
	// cluster are in maintenance mode.
	MaintenanceAnnotationLabel = "caps." + v1alpha2.GroupName + "/maintenance"
)
