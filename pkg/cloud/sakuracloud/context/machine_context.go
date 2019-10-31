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

package context

import (
	"context"
	"fmt"

	"sigs.k8s.io/cluster-api/util/patch"

	"github.com/sacloud/cluster-api-provider-sakuracloud/pkg/cloud/sakuracloud/session"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha2"

	infrav1 "github.com/sacloud/cluster-api-provider-sakuracloud/api/v1alpha2"

	clusterv1errors "sigs.k8s.io/cluster-api/errors"
)

// MachineContextParams are the parameters needed to create a MachineContext.
type MachineContextParams struct {
	ClusterContextParams
	Machine            *clusterv1.Machine
	SakuraCloudMachine *infrav1.SakuraCloudMachine
}

// MachineContext is a Go context used with a CAPI cluster.
type MachineContext struct {
	*ClusterContext
	Machine            *clusterv1.Machine
	SakuraCloudMachine *infrav1.SakuraCloudMachine
	Session            *session.Client
	patchHelper        *patch.Helper
}

// NewMachineContextFromClusterContext creates a new MachineContext using an
// existing CluserContext.
func NewMachineContextFromClusterContext(
	clusterCtx *ClusterContext,
	machine *clusterv1.Machine,
	sakuracloudMachine *infrav1.SakuraCloudMachine) (*MachineContext, error) {

	clusterCtx.Logger = clusterCtx.Logger.WithName(machine.Name)

	session, err := getOrCreateSession()
	if err != nil {
		return nil, err
	}

	helper, err := patch.NewHelper(sakuracloudMachine, clusterCtx.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	machineCtx := &MachineContext{
		ClusterContext:     clusterCtx,
		Machine:            machine,
		SakuraCloudMachine: sakuracloudMachine,
		Session:            session,
		patchHelper:        helper,
	}

	return machineCtx, nil
}

// NewMachineContext returns a new MachineContext.
func NewMachineContext(params *MachineContextParams) (*MachineContext, error) {
	ctx, err := NewClusterContext(&params.ClusterContextParams)
	if err != nil {
		return nil, err
	}
	return NewMachineContextFromClusterContext(ctx, params.Machine, params.SakuraCloudMachine)
}

// Strings returns ClusterNamespace/ClusterName/MachineName
func (c *MachineContext) String() string {
	if c.Machine == nil {
		return c.ClusterContext.String()
	}
	return fmt.Sprintf("%s/%s/%s", c.Cluster.Namespace, c.Cluster.Name, c.Machine.Name)
}

// GetObject returns the Machine object.
func (c *MachineContext) GetObject() runtime.Object {
	return c.Machine
}

// Zone returns the name of target zone.
func (c *MachineContext) Zone() string {
	return c.SakuraCloudCluster.Spec.Zone
}

// SetMachineError sets error details
func (c *MachineContext) SetMachineError(reason clusterv1errors.MachineStatusError, msg string) {
	c.SakuraCloudMachine.Status.ErrorReason = &reason
	c.SakuraCloudMachine.Status.ErrorMessage = &msg
}

// Patch updates the object and its status on the API server.
func (c *MachineContext) Patch() error {
	return c.patchHelper.Patch(context.TODO(), c.SakuraCloudMachine)
}
