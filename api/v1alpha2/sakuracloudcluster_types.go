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
	"sigs.k8s.io/cluster-api/errors"
)

const (
	// ClusterFinalizer allows ReconcileSakuraCloudCluster to clean up SakuraCloud
	// resources associated with SakuraCloudCluster before removing it from the
	// API server.
	ClusterFinalizer = "sakuracloudcluster.infrastructure.cluster.x-k8s.io"
)

// SakuraCloudClusterSpec defines the desired state of SakuraCloudCluster
type SakuraCloudClusterSpec struct {
	Zone                       string                    `json:"zone"`
	CloudProviderConfiguration SakuraCloudProviderConfig `json:"cloudProviderConfiguration,omitempty"`
}

// SakuraCloudClusterStatus defines the observed state of SakuraCloudClusterSpec
type SakuraCloudClusterStatus struct {
	Ready bool `json:"ready"`
	// APIEndpoints represents the endpoints to communicate with the control
	// plane.
	// +optional
	APIEndpoints []APIEndpoint `json:"apiEndpoints,omitempty"`

	// ErrorReason will be set in the event that there is a terminal problem
	// reconciling the Machine and will contain a succinct value suitable
	// for machine interpretation.
	//
	// This field should not be set for transitive errors that a controller
	// faces that are expected to be fixed automatically over
	// time (like service outages), but instead indicate that something is
	// fundamentally wrong with the Machine's spec or the configuration of
	// the controller, and that manual intervention is required. Examples
	// of terminal errors would be invalid combinations of settings in the
	// spec, values that are unsupported by the controller, or the
	// responsible controller itself being critically misconfigured.
	//
	// Any transient errors that occur during the reconciliation of Machines
	// can be added as events to the Machine object and/or logged in the
	// controller's output.
	// +optional
	ErrorReason *errors.ClusterStatusError `json:"errorReason,omitempty"`

	// ErrorMessage will be set in the event that there is a terminal problem
	// reconciling the Machine and will contain a more verbose string suitable
	// for logging and human consumption.
	//
	// This field should not be set for transitive errors that a controller
	// faces that are expected to be fixed automatically over
	// time (like service outages), but instead indicate that something is
	// fundamentally wrong with the Machine's spec or the configuration of
	// the controller, and that manual intervention is required. Examples
	// of terminal errors would be invalid combinations of settings in the
	// spec, values that are unsupported by the controller, or the
	// responsible controller itself being critically misconfigured.
	//
	// Any transient errors that occur during the reconciliation of Machines
	// can be added as events to the Machine object and/or logged in the
	// controller's output.
	// +optional
	ErrorMessage *string `json:"errorMessage,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=sakuracloudclusters,scope=Namespaced,categories=cluster-api
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Zone",type="string",JSONPath=".spec.zone",description="name of the SakuraCloud zone"
// +kubebuilder:printcolumn:name="Archive",type="string",JSONPath=".status.sourceArchive.name",description="name of the source archive"

// SakuraCloudCluster is the Schema for the sakuracloudclusters API
type SakuraCloudCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SakuraCloudClusterSpec   `json:"spec,omitempty"`
	Status SakuraCloudClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SakuraCloudClusterList contains a list of SakuraCloudCluster
type SakuraCloudClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SakuraCloudCluster `json:"items"`
}

// SakuraCloudProviderConfig is the Schema for the SakurCloudControllerManger configuration
//
// TODO 要修正
type SakuraCloudProviderConfig struct {
	// AccessToken .
	// +optional
	AccessToken string `json:"accessToken,omitempty"`

	// AccessSecret .
	// +optional
	AccessSecret string `json:"accessSecret,omtempty"`

	// Zone .
	// +optional
	Zone string `json:"zone,omitempty"`

	// Image .
	// +optional
	Image string `json:"image,omitempty"`

	// ClusterID .
	// +optional
	ClusterID string `json:"clusterID,omitempty"`
}

func init() {
	SchemeBuilder.Register(&SakuraCloudCluster{}, &SakuraCloudClusterList{})
}
