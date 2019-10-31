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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/errors"
)

const (
	// MachineFinalizer allows ReconcileSakuraCloudMachine to clean up SakuraCloud
	// resources associated with SakuraCloudMachine before removing it from the
	// API Server.
	MachineFinalizer = "sakuracloudmachine.infrastructure.cluster.x-k8s.io"
)

// SakuraCloudMachineSpec defines the desired state of SakuraCloudMachine
type SakuraCloudMachineSpec struct {

	// ProviderID is the unique identifier as specified by the cloud provider.
	ProviderID *string `json:"providerID,omitempty"`

	// This value is set automatically at runtime and should not be set or
	// modified by users.
	// MachineRef is used to lookup the VM.
	// +optional
	MachineRef *SakuraCloudResourceReference `json:"machineRef,omitempty"`

	// SourceArchive .
	SourceArchive SakuraCloudResourceReference `json:"sourceArchive"`

	// CPUs is the number of virtual processors in a virtual machine.
	// Defaults to the analogue property value in the template from which this
	// machine is cloned.
	// +optional
	CPUs int `json:"cpus,omitempty"`
	// MemoryMiB is the size of a virtual machine's memory, in GB.
	// +optional
	MemoryGB int `json:"memoryGB,omitempty"`
	// DiskGiB is the size of a virtual machine's disk, in GB.
	// +optional
	DiskGB int `json:"diskGB,omitempty"`
}

// SakuraCloudMachineStatus defines the observed state of SakuraCloudMachine
type SakuraCloudMachineStatus struct {
	// Ready is true when the provider resource is ready.
	// +optional
	Ready bool `json:"ready"`

	// Addresses contains the SakuraCloud instance associated addresses.
	Addresses []v1.NodeAddress `json:"addresses,omitempty"`

	// SourceArchiveInfo represents information of the node template image
	//
	// This value is set automatically at runtime and should not be set or
	// modified by users.
	// +optional
	SourceArchive *SourceArchiveInfo `json:"sourceArchive,omitempty"`

	// State is the state of the SakuraCloud instance for this machine.
	State InstanceState `json:"state,omitempty"`

	// JobRef is a managed object reference to a Job related to the
	// SakuraCloud resources.
	// This value is set automatically at runtime and should not be set or
	// modified by users.
	// +optional
	JobRef string `json:"jobRef,omitempty"`

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
	ErrorReason *errors.MachineStatusError `json:"errorReason,omitempty"`

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
// +kubebuilder:resource:path=sakuracloudmachines,scope=Namespaced,categories=cluster-api
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="CPUs",type="integer",JSONPath=".spec.cpus",description="number of CPUs"
// +kubebuilder:printcolumn:name="Memory",type="integer",JSONPath=".spec.memoryGB",description="size of memory"
// +kubebuilder:printcolumn:name="Disk",type="integer",JSONPath=".spec.diskGB",description="size of the disks"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.state",description="current status of the machine"

// SakuraCloudMachine is the Schema for the sakuracloudmachines API
type SakuraCloudMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SakuraCloudMachineSpec   `json:"spec,omitempty"`
	Status SakuraCloudMachineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SakuraCloudMachineList contains a list of SakuraCloudMachine
type SakuraCloudMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SakuraCloudMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SakuraCloudMachine{}, &SakuraCloudMachineList{})
}
