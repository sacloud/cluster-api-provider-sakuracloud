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
package session

import (
	"context"

	infrav1 "github.com/sacloud/cluster-api-provider-sakuracloud/api/v1alpha2"
	"github.com/sacloud/libsacloud/v2/sacloud"
	sacloudtypes "github.com/sacloud/libsacloud/v2/sacloud/types"
)

type ServerAPI interface {
	Read(ctx context.Context, zone string, serverID sacloudtypes.ID) (*sacloud.Server, error)
	Cleanup(ctx context.Context, zone string, serverID sacloudtypes.ID) JobID
	Provision(ctx context.Context, zone string, param *ServerBuildParameter) JobID
	FindArchive(ctx context.Context, zone string, filters []infrav1.Filter) (*sacloud.Archive, error)
	ReadArchive(ctx context.Context, zone string, archiveID sacloudtypes.ID) (*sacloud.Archive, error)
}

type ServerBuildParameter struct {
	ServerName      string
	ClusterName     string
	NameSpace       string
	IsControlPlane  bool
	SourceArchiveID string
	BootstrapData   string
	Spec            infrav1.SakuraCloudMachineSpec
}
