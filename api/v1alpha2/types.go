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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha2"
)

const (
	// AnnotationClusterInfrastructureReady indicates the cluster's
	// infrastructure sources are ready and machines may be created.
	AnnotationClusterInfrastructureReady = "sakuracloud.infrastructure.cluster.x-k8s.io/infrastructure-ready"

	// AnnotationControlPlaneReady indicates the cluster's control plane is
	// ready.
	AnnotationControlPlaneReady = "sakuracloud.infrastructure.cluster.x-k8s.io/control-plane-ready"

	// ValueReady is the ready value for *Ready annotations.
	ValueReady = "true"
)

// SourceArchiveInfo represents information of node template image
type SourceArchiveInfo struct {
	// ID .
	ID string `json:"id,omitempty"`
	// Name .
	Name string `json:"name,omitempty"`
}

// SakuraCloudResourceReference is a reference to a specific SakuraCloud resource by ID+Zone or filters.
// Only one of ID+Zone or Filters may be specified. Specifying more than one will result in
// a validation error.
type SakuraCloudResourceReference struct {
	// ID of resource
	// +optional
	ID *string `json:"id,omitempty"`

	// Filters is a set of key/value pairs used to identify a resource
	// They are applied according to the rules defined by the SakuraCloud API:
	// https://developer.sakura.ad.jp/cloud/api/1.1/
	//
	// If SakuraCloud API with Filters returns multiple results,
	// it use first data of results
	// +optional
	Filters []Filter `json:"filters,omitempty"`
}

// Filter is a filter used to identify an SakuraCloud resource
type Filter struct {
	// Name of the filter. Filter names are case-sensitive.
	Name string `json:"name"`

	// Values includes one or more filter values. Filter values are case-sensitive.
	Values []string `json:"values"`
}

// SakuraCloudMachineTemplateResource describes the data needed to create a SakuraCloudMachine from a template
type SakuraCloudMachineTemplateResource struct {
	metav1.TypeMeta `json:",inline"`

	// Standard object's metadata.
	clusterv1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the specification of the desired behavior of the machine.
	Spec SakuraCloudMachineSpec `json:"spec"`
}

// APIEndpoint represents a reachable Kubernetes API endpoint.
type APIEndpoint struct {
	// The hostname on which the API server is serving.
	Host string `json:"host"`

	// The port on which the API server is serving.
	Port int `json:"port"`
}

// InstanceState describe the state of an instance
type InstanceState string

const (
	// InstanceStatePending is the string representing an instance in pending state
	InstanceStatePending InstanceState = ""

	// InstanceStateProvisioning is the string representing an instance in ready state
	InstanceStateProvisioning = "provisioning"

	// InstanceStateReady is the string representing an instance in ready state
	InstanceStateReady = "ready"

	// InstanceStateCleaning is the string representing an instance in shutting-down state
	InstanceStateCleaning = "cleaning"

	// InstanceStateNotFound is the string representing an instance in not-found state
	InstanceStateNotFound = "notfound"
)
