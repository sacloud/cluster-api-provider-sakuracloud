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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	clusterutilv1 "sigs.k8s.io/cluster-api/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrav1 "github.com/sacloud/cluster-api-provider-sakuracloud/api/v1alpha2"
	"github.com/sacloud/cluster-api-provider-sakuracloud/pkg/cloud/sakuracloud/config"
	"github.com/sacloud/cluster-api-provider-sakuracloud/pkg/cloud/sakuracloud/context"
	"github.com/sacloud/cluster-api-provider-sakuracloud/pkg/cloud/sakuracloud/services/cloudprovider"
	infrautilv1 "github.com/sacloud/cluster-api-provider-sakuracloud/pkg/cloud/sakuracloud/util"
)

const (
	controllerName  = "sakuracloudcluster-controller"
	apiEndpointPort = 6443
)

// SakuraCloudClusterReconciler reconciles a SakuraCloudCluster object
//
// TODO Evnetへの対応
type SakuraCloudClusterReconciler struct {
	client.Client
	Recorder record.EventRecorder
	Log      logr.Logger
}

// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;patch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=sakuracloudclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=sakuracloudclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch
// +kubebuilder:rbac:groups=bootstrap.cluster.x-k8s.io,resources=kubeadmconfigs;kubeadmconfigs/status,verbs=get;list;watch

// Reconcile ensures the back-end state reflects the Kubernetes resource state intent.
func (r *SakuraCloudClusterReconciler) Reconcile(req ctrl.Request) (_ ctrl.Result, reterr error) {
	parentContext := goctx.Background()

	logger := r.Log.WithName(controllerName).
		WithName(fmt.Sprintf("namespace=%s", req.Namespace)).
		WithName(fmt.Sprintf("sakuracloudCluster=%s", req.Name))

	// Fetch the SakuraCloudCluster instance
	sakuracloudCluster := &infrav1.SakuraCloudCluster{}
	err := r.Get(parentContext, req.NamespacedName, sakuracloudCluster)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	logger = logger.WithName(sakuracloudCluster.APIVersion)

	// Fetch the Cluster.
	cluster, err := clusterutilv1.GetOwnerCluster(parentContext, r.Client, sakuracloudCluster.ObjectMeta)
	if err != nil {
		return reconcile.Result{}, err
	}
	if cluster == nil {
		logger.Info("Waiting for Cluster Controller to set OwnerRef on SakuraCloudCluster")
		return reconcile.Result{RequeueAfter: config.DefaultRequeue}, nil
	}

	logger = logger.WithName(fmt.Sprintf("cluster=%s", cluster.Name))

	// Create the context.
	ctx, err := context.NewClusterContext(&context.ClusterContextParams{
		Context:            parentContext,
		Cluster:            cluster,
		SakuraCloudCluster: sakuracloudCluster,
		Client:             r.Client,
		Logger:             logger,
	})
	if err != nil {
		return reconcile.Result{}, errors.Errorf("failed to create cluster context: %+v", err)
	}

	// Always close the context when exiting this function so we can persist any SakuraCloudCluster changes.
	defer func() {
		if err := ctx.Patch(); err != nil && reterr == nil {
			reterr = err
		}
	}()

	// Handle deleted clusters
	if !sakuracloudCluster.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx)
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(ctx)
}

func (r *SakuraCloudClusterReconciler) reconcileDelete(ctx *context.ClusterContext) (reconcile.Result, error) {
	ctx.Logger.Info("Reconciling SakuraCloudCluster delete")

	// Cluster is deleted so remove the finalizer.
	ctx.SakuraCloudCluster.Finalizers = clusterutilv1.Filter(ctx.SakuraCloudCluster.Finalizers, infrav1.ClusterFinalizer)

	return reconcile.Result{}, nil
}

func (r *SakuraCloudClusterReconciler) reconcileNormal(ctx *context.ClusterContext) (reconcile.Result, error) {
	ctx.Logger.Info("Reconciling SakuraCloudCluster")

	ctx.SakuraCloudCluster.Status.Ready = true
	ctx.Logger.V(6).Info("SakuraCloudCluster is infrastructure-ready")

	// If the SakuraCloudCluster doesn't have our finalizer, add it.
	if !clusterutilv1.Contains(ctx.SakuraCloudCluster.Finalizers, infrav1.ClusterFinalizer) {
		ctx.SakuraCloudCluster.Finalizers = append(ctx.SakuraCloudCluster.Finalizers, infrav1.ClusterFinalizer)
		ctx.Logger.V(6).Info(
			"adding finalizer for SakuraCloudCluster",
			"cluster-namespace", ctx.SakuraCloudCluster.Namespace,
			"cluster-name", ctx.SakuraCloudCluster.Name)
	}

	// Update the SakuraCloudCluster resource with its API enpoints.
	if err := r.reconcileAPIEndpoints(ctx); err != nil {
		return reconcile.Result{}, errors.Wrapf(err,
			"failed to reconcile API endpoints for SakuraCloudCluster %s/%s",
			ctx.SakuraCloudCluster.Namespace, ctx.SakuraCloudCluster.Name)
	}

	// Create the external cloud provider addons
	if err := r.reconcileCloudProvider(ctx); err != nil {
		return reconcile.Result{}, errors.Wrapf(err,
			"failed to reconcile cloud provider for SakuraCloudCluster %s/%s",
			ctx.SakuraCloudCluster.Namespace, ctx.SakuraCloudCluster.Name)
	}

	return reconcile.Result{}, nil
}

func (r *SakuraCloudClusterReconciler) reconcileAPIEndpoints(ctx *context.ClusterContext) error {
	// If the cluster already has API endpoints set then there is nothing to do.
	if len(ctx.SakuraCloudCluster.Status.APIEndpoints) > 0 {
		ctx.Logger.V(6).Info("API endpoints already exist")
		return nil
	}

	// Get the CAPI Machine resources for the cluster.
	machines, err := infrautilv1.GetMachinesInCluster(ctx, ctx.Client, ctx.SakuraCloudCluster.Namespace, ctx.SakuraCloudCluster.Name)
	if err != nil {
		return errors.Wrapf(err,
			"failed to get Machines for Cluster %s/%s",
			ctx.SakuraCloudCluster.Namespace, ctx.SakuraCloudCluster.Name)
	}

	// Iterate over the cluster's control plane CAPI machines.
	for _, machine := range clusterutilv1.GetControlPlaneMachines(machines) {

		// TODO HA対応
		var apiEndpoint infrav1.APIEndpoint

		if machine.Spec.Bootstrap.Data == nil {
			ctx.Logger.V(6).Info(
				"skipping machine while looking for IP address",
				"machine-name", machine.Name,
				"skip-reason", "nilBootstrapData")
			continue
		}

		// Get the SakuraCloudMachine for the CAPI Machine resource.
		sakuracloudMachine, err := infrautilv1.GetSakuraCloudMachine(ctx, ctx.Client, machine.Namespace, machine.Name)
		if err != nil {
			return errors.Wrapf(err,
				"failed to get SakuraCloudMachine for Machine %s/%s/%s",
				machine.Namespace, ctx.SakuraCloudCluster.Name, machine.Name)
		}

		// Get the SakuraCloudMachine's preferred IP address.
		ipAddr, err := infrautilv1.GetMachinePreferredIPAddress(sakuracloudMachine)
		if err != nil {
			if err == infrautilv1.ErrNoMachineIPAddr {
				continue
			}
			return errors.Wrapf(err,
				"failed to get preferred IP address for SakuraCloudMachine %s/%s/%s",
				machine.Namespace, ctx.SakuraCloudCluster.Name, sakuracloudMachine.Name)
		}

		apiEndpoint.Host = ipAddr
		apiEndpoint.Port = apiEndpointPort

		ctx.Logger.V(6).Info(
			"found API endpoint via control plane machine",
			"host", apiEndpoint.Host, "port", apiEndpoint.Port)

		// Set APIEndpoints so the CAPI controller can read the API endpoints
		// for this SakuraCloudCluster into the analogous CAPI Cluster using an
		// UnstructuredReader.
		ctx.SakuraCloudCluster.Status.APIEndpoints = []infrav1.APIEndpoint{apiEndpoint}
		return nil
	}
	return infrautilv1.ErrNoMachineIPAddr
}

// SetupWithManager adds this controller to the provided manager.
func (r *SakuraCloudClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.SakuraCloudCluster{}).
		Complete(r)
}

func (r *SakuraCloudClusterReconciler) reconcileCloudProvider(ctx *context.ClusterContext) error {
	// if the cloud provider image is not specified, then we do nothing
	conf := ctx.SakuraCloudCluster.Spec.CloudProviderConfiguration

	// TODO Webhookへ移動しValidationも追加する
	if conf.Image == "" {
		conf.Image = "sacloud/sakura-cloud-controller-manager:latest"
	}
	if conf.AccessToken == "" {
		conf.AccessToken = ctx.AccessToken()
	}
	if conf.AccessSecret == "" {
		conf.AccessSecret = ctx.AccessSecret()
	}
	if conf.Zone == "" {
		conf.Zone = ctx.Zone()
	}
	if conf.ClusterID == "" {
		conf.ClusterID = "sakuracloud"
	}

	targetClusterClient, err := infrautilv1.NewKubeClient(ctx, ctx.Client, ctx.Cluster)
	if err != nil {
		return errors.Wrapf(err,
			"failed to get client for Cluster %s/%s",
			ctx.Cluster.Namespace, ctx.Cluster.Name)
	}

	serviceAccount := cloudprovider.CloudControllerManagerServiceAccount()
	if _, err := targetClusterClient.CoreV1().ServiceAccounts(serviceAccount.Namespace).Create(serviceAccount); err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	clusterRole := cloudprovider.CloudControllerManagerClusterRole()
	if _, err := targetClusterClient.RbacV1().ClusterRoles().Create(clusterRole); err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	roleBinding := cloudprovider.CloudControllerManagerRoleBinding()
	if _, err := targetClusterClient.RbacV1().RoleBindings(roleBinding.Namespace).Create(roleBinding); err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	clusterRoleBinding := cloudprovider.CloudControllerManagerClusterRoleBinding()
	if _, err := targetClusterClient.RbacV1().ClusterRoleBindings().Create(clusterRoleBinding); err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	secret := cloudprovider.CloudControllerManagerCredential(conf.AccessToken, conf.AccessSecret)
	if _, err := targetClusterClient.CoreV1().Secrets(secret.Namespace).Create(secret); err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	deployment := cloudprovider.CloudControllerManagerDeployment(conf.Image, conf.Zone, conf.ClusterID)
	if _, err := targetClusterClient.AppsV1().Deployments(deployment.Namespace).Create(deployment); err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	return nil
}
