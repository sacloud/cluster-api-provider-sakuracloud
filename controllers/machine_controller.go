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

package controllers

import (
	goctx "context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/sacloud/cluster-api-provider-sakuracloud/pkg/cloud/sakuracloud/services"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha2"
	clusterutilv1 "sigs.k8s.io/cluster-api/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	infrav1 "github.com/sacloud/cluster-api-provider-sakuracloud/api/v1alpha2"
	"github.com/sacloud/cluster-api-provider-sakuracloud/pkg/cloud/sakuracloud/config"
	"github.com/sacloud/cluster-api-provider-sakuracloud/pkg/cloud/sakuracloud/context"

	clusterv1errors "sigs.k8s.io/cluster-api/errors"
)

const (
	machineControllerName = "sakuracloudmachine-controller"
)

// SakuraCloudMachineReconciler reconciles a SakuraCloudMachine object
//
// TODO Evnetへの対
type SakuraCloudMachineReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=sakuracloudmachines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=sakuracloudmachines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machines;machines/status,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch

// Reconcile ensures the back-end state reflects the Kubernetes resource state intent.
func (r *SakuraCloudMachineReconciler) Reconcile(req ctrl.Request) (_ ctrl.Result, reterr error) {
	parentContext := goctx.Background()

	logger := r.Log.
		WithName(machineControllerName).
		WithName(fmt.Sprintf("namespace=%s", req.Namespace)).
		WithName(fmt.Sprintf("sakuracloudMachine=%s", req.Name))

	// Fetch the SakuraCloudMachine instance.
	sakuracloudMachine := &infrav1.SakuraCloudMachine{}
	if err := r.Get(parentContext, req.NamespacedName, sakuracloudMachine); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	logger = logger.WithName(sakuracloudMachine.APIVersion)

	// Fetch the Machine.
	machine, err := clusterutilv1.GetOwnerMachine(parentContext, r.Client, sakuracloudMachine.ObjectMeta)
	if err != nil {
		return reconcile.Result{}, err
	}
	if machine == nil {
		logger.Info("Waiting for Machine Controller to set OwnerRef on SakuraCloudMachine")
		return reconcile.Result{RequeueAfter: config.DefaultRequeue}, nil
	}

	logger = logger.WithName(fmt.Sprintf("machine=%s", machine.Name))

	// Fetch the Cluster.
	cluster, err := clusterutilv1.GetClusterFromMetadata(parentContext, r.Client, machine.ObjectMeta)
	if err != nil {
		logger.Info("Machine is missing cluster label or cluster does not exist")
		return reconcile.Result{}, nil
	}

	logger = logger.WithName(fmt.Sprintf("cluster=%s", cluster.Name))

	sakuracloudCluster := &infrav1.SakuraCloudCluster{}

	sakuracloudClusterName := client.ObjectKey{
		Namespace: sakuracloudMachine.Namespace,
		Name:      cluster.Spec.InfrastructureRef.Name,
	}
	if err := r.Client.Get(parentContext, sakuracloudClusterName, sakuracloudCluster); err != nil {
		logger.Info("Waiting for SakuraCloudCluster")
		return reconcile.Result{RequeueAfter: config.DefaultRequeue}, nil
	}

	logger = logger.WithName(fmt.Sprintf("sakuracloudCluster=%s", sakuracloudCluster.Name))

	// Create the cluster context.
	clusterContext, err := context.NewClusterContext(&context.ClusterContextParams{
		Context:            parentContext,
		Cluster:            cluster,
		SakuraCloudCluster: sakuracloudCluster,
		Client:             r.Client,
		Logger:             logger,
	})
	if err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to create cluster context")
	}

	// Create the machine context
	machineContext, err := context.NewMachineContextFromClusterContext(
		clusterContext,
		machine,
		sakuracloudMachine)
	if err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to create machine context")
	}

	// complete cluster spec
	if sakuracloudMachine.Spec.SourceArchive.ID == nil {
		archive, err := machineContext.Session.FindArchive(machineContext, machineContext.Zone(), sakuracloudMachine.Spec.SourceArchive.Filters)
		if err != nil {
			machineContext.SetMachineError(clusterv1errors.InvalidConfigurationMachineError, err.Error())
			return reconcile.Result{}, errors.Errorf("failed to set source archive id: %+v", err)
		}

		if archive == nil {
			machineContext.SetMachineError(clusterv1errors.InvalidConfigurationMachineError, "archive not found")
			return reconcile.Result{}, errors.Errorf("failed to set source archive id: %+v", "archive not found")
		}

		id := archive.ID.String()
		machineContext.SakuraCloudMachine.Spec.SourceArchive.ID = &id
	}

	if sakuracloudMachine.Status.SourceArchive == nil {
		archive, err := machineContext.Session.ReadArchive(machineContext, machineContext.Zone(), types.StringID(*sakuracloudMachine.Spec.SourceArchive.ID))
		if err != nil {
			machineContext.SetMachineError(clusterv1errors.InvalidConfigurationMachineError, err.Error())
			return reconcile.Result{}, errors.Errorf("failed to get source archive info: %+v", err)
		}
		sakuracloudMachine.Status.SourceArchive = &infrav1.SourceArchiveInfo{
			ID:   archive.ID.String(),
			Name: archive.Name,
		}
	}

	// Always close the context when exiting this function so we can persist any SakuraCloudMachine changes.
	defer func() {
		if err := machineContext.Patch(); err != nil && reterr == nil {
			reterr = err
		}
	}()

	// Handle deleted machines
	if !sakuracloudMachine.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(machineContext)
	}

	// Handle non-deleted machines
	return r.reconcileNormal(machineContext)
}

// SetupWithManager adds this controller to the provided manager.
func (r *SakuraCloudMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.SakuraCloudMachine{}).Watches(
		&source.Kind{Type: &clusterv1.Machine{}},
		&handler.EnqueueRequestsFromMapFunc{
			ToRequests: clusterutilv1.MachineToInfrastructureMapFunc(schema.GroupVersionKind{
				Group:   infrav1.SchemeBuilder.GroupVersion.Group,
				Version: infrav1.SchemeBuilder.GroupVersion.Version,
				Kind:    "SakuraCloudMachine",
			}),
		},
	).Complete(r)
}

func (r *SakuraCloudMachineReconciler) reconcileDelete(ctx *context.MachineContext) (reconcile.Result, error) {
	ctx.Logger.Info("Handling deleted SakuraCloudMachine")

	var service services.SakuraCloudMachineInterface = &services.SakuraCloudService{}

	server, err := service.DestroyServer(ctx)
	if err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to destroy server")
	}

	// Requeue the operation until the VM is "notfound".
	if server.Status.State != infrav1.InstanceStateNotFound {
		ctx.Logger.V(6).Info("requeuing operation until server state is reconciled", "expected-state", infrav1.InstanceStateNotFound, "actual-state", server.Status.State)
		return reconcile.Result{RequeueAfter: config.DefaultRequeue}, nil
	}

	// The server is deleted so remove the finalizer.
	ctx.SakuraCloudMachine.ObjectMeta.Finalizers = clusterutilv1.Filter(ctx.SakuraCloudMachine.Finalizers, infrav1.MachineFinalizer)
	return reconcile.Result{}, nil
}

func (r *SakuraCloudMachineReconciler) reconcileNormal(ctx *context.MachineContext) (reconcile.Result, error) {
	// If the SakuraCloudMachine is in an error state, return early.
	if ctx.SakuraCloudMachine.Status.ErrorReason != nil || ctx.SakuraCloudMachine.Status.ErrorMessage != nil {
		ctx.Logger.Info("Error state detected, skipping reconciliation")
		return reconcile.Result{}, nil
	}

	// If the SakuraCloudMachine doesn't have our finalizer, add it.
	if !clusterutilv1.Contains(ctx.SakuraCloudMachine.Finalizers, infrav1.MachineFinalizer) {
		ctx.SakuraCloudMachine.Finalizers = append(ctx.SakuraCloudMachine.Finalizers, infrav1.MachineFinalizer)
	}

	if !ctx.Cluster.Status.InfrastructureReady {
		ctx.Logger.Info("Cluster infrastructure is not ready yet, requeuing machine")
		return reconcile.Result{RequeueAfter: config.DefaultRequeue}, nil
	}

	// Make sure bootstrap data is available and populated.
	if ctx.Machine.Spec.Bootstrap.Data == nil {
		ctx.Logger.Info("Waiting for bootstrap data to be available")
		return reconcile.Result{RequeueAfter: config.DefaultRequeue}, nil
	}

	var service services.SakuraCloudMachineInterface = &services.SakuraCloudService{}
	sacloudMachine, err := service.ReconcileServer(ctx)
	if err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to reconcile VM")
	}

	if sacloudMachine.Status.State != infrav1.InstanceStateReady {
		ctx.Logger.V(6).Info("requeuing operation until vm state is reconciled", "expected-vm-state", infrav1.InstanceStateReady, "actual-vm-state", sacloudMachine.Status.State)
		return reconcile.Result{RequeueAfter: config.DefaultRequeue}, nil
	}

	if err := r.reconcileProviderID(ctx, sacloudMachine, service); err != nil {
		return reconcile.Result{RequeueAfter: config.DefaultRequeue}, err
	}

	ctx.SakuraCloudMachine.Status.Ready = true
	ctx.Logger.V(6).Info("SakuraCloudMachine is infrastructure-ready")

	return reconcile.Result{}, nil
}

func (r *SakuraCloudMachineReconciler) reconcileProviderID(ctx *context.MachineContext, sacloudMachine *infrav1.SakuraCloudMachine, service services.SakuraCloudMachineInterface) error {
	providerID := fmt.Sprintf("sakuracloud://%s", *sacloudMachine.Spec.MachineRef.ID)
	if ctx.SakuraCloudMachine.Spec.ProviderID == nil || *ctx.SakuraCloudMachine.Spec.ProviderID != providerID {
		ctx.SakuraCloudMachine.Spec.ProviderID = &providerID
		ctx.Logger.V(6).Info("updated provider ID", "provider-id", providerID)
	}
	return nil
}
