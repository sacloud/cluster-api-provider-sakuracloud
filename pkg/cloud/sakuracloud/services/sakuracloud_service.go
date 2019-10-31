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
package services

import (
	"github.com/sacloud/cluster-api-provider-sakuracloud/pkg/cloud/sakuracloud/session"

	infrav1 "github.com/sacloud/cluster-api-provider-sakuracloud/api/v1alpha2"
	"github.com/sacloud/cluster-api-provider-sakuracloud/pkg/cloud/sakuracloud/context"
	"github.com/sacloud/cluster-api-provider-sakuracloud/pkg/cloud/sakuracloud/util"
	sacloudtypes "github.com/sacloud/libsacloud/v2/sacloud/types"
	"sigs.k8s.io/cluster-api/errors"

	corev1 "k8s.io/api/core/v1"
)

// SakuraCloudService is a service for creating/updating/deleting virtual
// machines on SakuraCloud
type SakuraCloudService struct{}

// ReconcileVM reconciles a VM with the intended state
func (s *SakuraCloudService) ReconcileServer(ctx *context.MachineContext) (*infrav1.SakuraCloudMachine, error) {
	// TODO Updateの考慮

	if ctx.SakuraCloudMachine.Status.State == infrav1.InstanceStateReady {
		return ctx.SakuraCloudMachine, nil
	}

	// If there is no pending task or no machine ref then no VM exits, create one
	if ctx.SakuraCloudMachine.Status.State == infrav1.InstanceStatePending && ctx.SakuraCloudMachine.Status.JobRef == "" {

		jobID := ctx.Session.Provision(ctx, ctx.Zone(), &session.ServerBuildParameter{
			ServerName:      ctx.Machine.Name,
			ClusterName:     ctx.Cluster.Name,
			NameSpace:       ctx.Cluster.Namespace,
			IsControlPlane:  util.IsControlPlaneMachine(ctx.Machine),
			SourceArchiveID: ctx.SakuraCloudMachine.Status.SourceArchive.ID,
			BootstrapData:   *ctx.Machine.Spec.Bootstrap.Data,
			Spec:            ctx.SakuraCloudMachine.Spec,
		})
		ctx.SakuraCloudMachine.Status.JobRef = string(jobID)
		ctx.SakuraCloudMachine.Status.State = infrav1.InstanceStateProvisioning
		return ctx.SakuraCloudMachine, nil
	}

	job := ctx.Session.JobByID(ctx.SakuraCloudMachine.Status.JobRef)
	if job == nil {
		return ctx.SakuraCloudMachine, nil // waiting for start
	}

	if job.Error != nil {
		ctx.SetMachineError(errors.CreateMachineError, job.Error.Error())
		return ctx.SakuraCloudMachine, job.Error
	}

	if job.Type != session.JobTypeProvisioning {
		return ctx.SakuraCloudMachine, nil
	}

	if job.Reference != nil && !job.Reference.ServerID.IsEmpty() {
		id := job.Reference.ServerID.String()
		ctx.SakuraCloudMachine.Spec.MachineRef = &infrav1.SakuraCloudResourceReference{
			ID: &id,
		}

		sv, err := ctx.Session.Read(ctx, ctx.Zone(), job.Reference.ServerID)
		if err != nil || sv == nil {
			return ctx.SakuraCloudMachine, err
		}

		// TODO あとで修正
		ctx.SakuraCloudMachine.Status.Addresses = []corev1.NodeAddress{
			{
				Type:    corev1.NodeExternalIP,
				Address: sv.Interfaces[0].IPAddress,
			},
		}
	}

	switch job.State {
	case session.JobStatePending, session.JobStateInFlight:
		return ctx.SakuraCloudMachine, nil
	case session.JobStateFailed:
		ctx.SetMachineError(errors.CreateMachineError, job.Error.Error())
		return ctx.SakuraCloudMachine, job.Error
	case session.JobStateDone:
		ctx.SakuraCloudMachine.Status.JobRef = ""
		ctx.Session.DeleteJob(string(job.ID))
		ctx.SakuraCloudMachine.Status.State = infrav1.InstanceStateReady
	}

	return ctx.SakuraCloudMachine, nil
}

// DestroyVM powers off and removes a VM from the inventory
func (s *SakuraCloudService) DestroyServer(ctx *context.MachineContext) (*infrav1.SakuraCloudMachine, error) {
	if ctx.SakuraCloudMachine.Status.State == infrav1.InstanceStateNotFound {
		return ctx.SakuraCloudMachine, nil
	}

	if ctx.SakuraCloudMachine.Status.State != infrav1.InstanceStateCleaning && ctx.SakuraCloudMachine.Status.JobRef == "" {
		if ctx.SakuraCloudMachine.Spec.MachineRef == nil {
			// server already deleted
			ctx.SakuraCloudMachine.Status.State = infrav1.InstanceStateNotFound
			return ctx.SakuraCloudMachine, nil
		}

		serverID := sacloudtypes.StringID(*ctx.SakuraCloudMachine.Spec.MachineRef.ID)
		jobID := ctx.Session.Cleanup(ctx, ctx.Zone(), serverID)

		ctx.SakuraCloudMachine.Status.JobRef = string(jobID)
		ctx.SakuraCloudMachine.Status.State = infrav1.InstanceStateCleaning
		return ctx.SakuraCloudMachine, nil
	}

	job := ctx.Session.JobByID(ctx.SakuraCloudMachine.Status.JobRef)
	if job == nil {
		return ctx.SakuraCloudMachine, nil
	}
	if job.Type != session.JobTypeCleaning {
		// cleanup old job and requeue
		ctx.SakuraCloudMachine.Status.JobRef = ""
		ctx.Session.DeleteJob(string(job.ID))
		return ctx.SakuraCloudMachine, nil
	}
	if job.Error != nil {
		ctx.SetMachineError(errors.DeleteMachineError, job.Error.Error())
		return ctx.SakuraCloudMachine, job.Error
	}

	switch job.State {
	case session.JobStatePending, session.JobStateInFlight:
		return ctx.SakuraCloudMachine, nil
	case session.JobStateFailed:
		ctx.SetMachineError(errors.CreateMachineError, job.Error.Error())
		return ctx.SakuraCloudMachine, job.Error
	case session.JobStateDone:
		ctx.SakuraCloudMachine.Spec.MachineRef = nil

		ctx.SakuraCloudMachine.Status.JobRef = ""
		ctx.Session.DeleteJob(string(job.ID))
		ctx.SakuraCloudMachine.Status.State = infrav1.InstanceStateNotFound
	}

	return ctx.SakuraCloudMachine, nil
}
