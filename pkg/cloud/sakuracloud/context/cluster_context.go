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
	"os"

	"sigs.k8s.io/cluster-api/util/patch"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/klogr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sacloud/cluster-api-provider-sakuracloud/api/v1alpha2"
	"github.com/sacloud/cluster-api-provider-sakuracloud/pkg/cloud/sakuracloud/session"

	clusterv1errors "sigs.k8s.io/cluster-api/errors"
)

// ClusterContextParams are the parameters needed to create a ClusterContext.
type ClusterContextParams struct {
	Context            context.Context
	Cluster            *clusterv1.Cluster
	SakuraCloudCluster *v1alpha2.SakuraCloudCluster
	Client             client.Client
	Logger             logr.Logger
}

// ClusterContext is a Go context used with a CAPI cluster.
type ClusterContext struct {
	context.Context
	Cluster            *clusterv1.Cluster
	SakuraCloudCluster *v1alpha2.SakuraCloudCluster
	Client             client.Client
	Logger             logr.Logger
	Session            *session.Client
	patchHelper        *patch.Helper
}

// NewClusterContext returns a new ClusterContext.
func NewClusterContext(params *ClusterContextParams) (*ClusterContext, error) {
	parentContext := params.Context
	if parentContext == nil {
		parentContext = context.Background()
	}

	logr := params.Logger
	if logr == nil {
		logr = klogr.New().WithName("default-logger")
	}
	logr = logr.WithName(params.Cluster.APIVersion).WithName(params.Cluster.Namespace).WithName(params.Cluster.Name)

	session, err := getOrCreateSession()
	if err != nil {
		return nil, err
	}
	helper, err := patch.NewHelper(params.SakuraCloudCluster, params.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	return &ClusterContext{
		Context:            parentContext,
		Cluster:            params.Cluster,
		SakuraCloudCluster: params.SakuraCloudCluster,
		Client:             params.Client,
		Logger:             logr,
		Session:            session,
		patchHelper:        helper,
	}, nil
}

// Strings returns ClusterNamespace/ClusterName
func (c *ClusterContext) String() string {
	return fmt.Sprintf("%s/%s", c.Cluster.Namespace, c.Cluster.Name)
}

// GetCluster returns the Cluster object.
func (c *ClusterContext) GetCluster() *clusterv1.Cluster {
	return c.Cluster
}

// GetClient returns the controller client.
func (c *ClusterContext) GetClient() client.Client {
	return c.Client
}

// GetObject returns the Cluster object.
func (c *ClusterContext) GetObject() runtime.Object {
	return c.Cluster
}

// GetLogger returns the Logger.
func (c *ClusterContext) GetLogger() logr.Logger {
	return c.Logger
}

// ClusterName returns the name of the cluster.
func (c *ClusterContext) ClusterName() string {
	return c.Cluster.Name
}

// AccessToken returns the username used to access the SakuraCloud API.
func (c *ClusterContext) AccessToken() string {
	return os.Getenv("SAKURACLOUD_ACCESS_TOKEN")
}

// AccessSecret returns the password used to access the SakuraCloud API.
func (c *ClusterContext) AccessSecret() string {
	return os.Getenv("SAKURACLOUD_ACCESS_TOKEN_SECRET")
}

// Zone returns the name of target zone.
func (c *ClusterContext) Zone() string {
	return c.SakuraCloudCluster.Spec.Zone
}

// SetClusterError sets error details
func (c *ClusterContext) SetClusterError(reason clusterv1errors.ClusterStatusError, msg string) {
	c.SakuraCloudCluster.Status.ErrorReason = &reason
	c.SakuraCloudCluster.Status.ErrorMessage = &msg
}

// Patch updates the object and its status on the API server.
func (c *ClusterContext) Patch() error {
	return c.patchHelper.Patch(context.TODO(), c.SakuraCloudCluster)
}
