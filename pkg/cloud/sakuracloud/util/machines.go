/*
Copyright 2019 The Kubernetes Authors.

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

package util

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha2"
	clusterutilv1 "sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github.com/sacloud/cluster-api-provider-sakuracloud/api/v1alpha2"
)

// GetMachinesInCluster gets a cluster's Machine resources.
func GetMachinesInCluster(
	ctx context.Context,
	controllerClient client.Client,
	namespace, clusterName string) ([]*clusterv1.Machine, error) {

	labels := map[string]string{clusterv1.MachineClusterLabelName: clusterName}
	machineList := &clusterv1.MachineList{}

	if err := controllerClient.List(
		ctx, machineList,
		client.InNamespace(namespace),
		client.MatchingLabels(labels)); err != nil {
		return nil, errors.Wrapf(
			err, "error getting machines in cluster %s/%s",
			namespace, clusterName)
	}

	machines := make([]*clusterv1.Machine, len(machineList.Items))
	for i := range machineList.Items {
		machines[i] = &machineList.Items[i]
	}

	return machines, nil
}

// GetSakuraCloudMachinesInCluster gets a cluster's SakuraCloudMachine resources.
func GetSakuraCloudMachinesInCluster(
	ctx context.Context,
	controllerClient client.Client,
	namespace, clusterName string) ([]*infrav1.SakuraCloudMachine, error) {

	labels := map[string]string{clusterv1.MachineClusterLabelName: clusterName}
	machineList := &infrav1.SakuraCloudMachineList{}

	if err := controllerClient.List(
		ctx, machineList,
		client.InNamespace(namespace),
		client.MatchingLabels(labels)); err != nil {
		return nil, err
	}

	machines := make([]*infrav1.SakuraCloudMachine, len(machineList.Items))
	for i := range machineList.Items {
		machines[i] = &machineList.Items[i]
	}

	return machines, nil
}

// GetSakuraCloudMachine gets a SakuraCloudMachine resource for the given CAPI Machine.
func GetSakuraCloudMachine(
	ctx context.Context,
	controllerClient client.Client,
	namespace, machineName string) (*infrav1.SakuraCloudMachine, error) {

	machine := &infrav1.SakuraCloudMachine{}
	namespacedName := apitypes.NamespacedName{
		Namespace: namespace,
		Name:      machineName,
	}
	if err := controllerClient.Get(ctx, namespacedName, machine); err != nil {
		return nil, err
	}
	return machine, nil
}

// ErrNoMachineIPAddr indicates that no valid IP addresses were found in a machine context
var ErrNoMachineIPAddr = errors.New("no IP addresses found for machine")

// GetMachinePreferredIPAddress returns the preferred IP address for a
// SakuraCloudMachine resource.
func GetMachinePreferredIPAddress(machine *infrav1.SakuraCloudMachine) (string, error) {
	for _, nodeAddr := range machine.Status.Addresses {
		if nodeAddr.Type != corev1.NodeExternalIP {
			//if nodeAddr.Type != corev1.NodeInternalIP {
			continue
		}
		return nodeAddr.Address, nil
	}

	return "", ErrNoMachineIPAddr
}

// IsControlPlaneMachine returns a flag indicating whether or not a machine has
// the control plane role.
func IsControlPlaneMachine(machine *clusterv1.Machine) bool {
	return clusterutilv1.IsControlPlaneMachine(machine)
}
