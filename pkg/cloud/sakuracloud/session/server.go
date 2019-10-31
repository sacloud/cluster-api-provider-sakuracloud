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
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/sacloud/libsacloud/v2/sacloud/search"

	"github.com/sacloud/ftps"
	"github.com/sacloud/libsacloud/v2/sacloud"
	sacloudtypes "github.com/sacloud/libsacloud/v2/sacloud/types"
	"github.com/sacloud/libsacloud/v2/utils/server"

	infrav1 "github.com/sacloud/cluster-api-provider-sakuracloud/api/v1alpha2"
)

type serverClient struct {
	caller sacloud.APICaller
	jobs   *jobRegistry
}

func (s *serverClient) serverOp() sacloud.ServerAPI {
	return sacloud.NewServerOp(s.caller)
}

func (s *serverClient) isoImageOp() sacloud.CDROMAPI {
	return sacloud.NewCDROMOp(s.caller)
}

func (s *serverClient) archiveOp() sacloud.ArchiveAPI {
	return sacloud.NewArchiveOp(s.caller)
}

func (s *serverClient) ReadArchive(ctx context.Context, zone string, archiveID sacloudtypes.ID) (*sacloud.Archive, error) {
	return s.archiveOp().Read(ctx, zone, archiveID)
}

func (s *serverClient) FindArchive(ctx context.Context, zone string, filters []infrav1.Filter) (*sacloud.Archive, error) {
	condition := &sacloud.FindCondition{
		Filter: search.Filter{},
	}
	for _, filter := range filters {
		condition.Filter[search.Key(filter.Name)] = search.ExactMatch(filter.Values...)
	}

	searched, err := s.archiveOp().Find(ctx, zone, condition)
	if err != nil {
		return nil, err
	}
	if searched.Count == 0 {
		return nil, nil
	}
	return searched.Archives[0], nil
}

func (s *serverClient) Read(ctx context.Context, zone string, id sacloudtypes.ID) (*sacloud.Server, error) {
	return s.serverOp().Read(ctx, zone, id)
}

func (s *serverClient) Cleanup(ctx context.Context, zone string, serverID sacloudtypes.ID) JobID {
	jobID := JobID(fmt.Sprintf("cleanup/%s/%s", zone, serverID))
	status := &JobStatus{
		ID:    jobID,
		Type:  JobTypeCleaning,
		State: JobStatePending,
		Reference: &CloudObjectRef{
			ServerID: serverID,
		},
	}
	s.jobs.set(jobID, status)

	go func() {
		status.State = JobStateInFlight

		sv, err := s.serverOp().Read(ctx, zone, serverID)
		if err != nil {
			if sacloud.IsNotFoundError(err) {
				status.State = JobStateDone
				return
			}
			status.Error = err
			status.State = JobStateFailed
			return
		}

		// shutdown
		if sv.InstanceStatus.IsUp() {
			if err := s.serverOp().Shutdown(ctx, zone, serverID, &sacloud.ShutdownOption{Force: true}); err != nil {
				status.Error = err
				status.State = JobStateFailed
				return
			}

			if _, err := sacloud.WaiterForDown(func() (interface{}, error) {
				return s.serverOp().Read(ctx, zone, serverID)
			}).WaitForState(ctx); err != nil {
				status.Error = err
				status.State = JobStateFailed
				return
			}
		}

		// delete server+disks
		var diskIDs []sacloudtypes.ID
		for _, disk := range sv.Disks {
			diskIDs = append(diskIDs, disk.ID)
		}
		if err := s.serverOp().DeleteWithDisks(ctx, zone, serverID, &sacloud.ServerDeleteWithDisksRequest{IDs: diskIDs}); err != nil {
			status.Error = err
			status.State = JobStateFailed
			return
		}

		// delete iso-image
		if !sv.CDROMID.IsEmpty() {
			if err := s.isoImageOp().Delete(ctx, zone, sv.CDROMID); err != nil {
				status.Error = err
				status.State = JobStateFailed
				return
			}
		}

		status.State = JobStateDone
	}()

	return jobID
}

func (s *serverClient) Provision(ctx context.Context, zone string, param *ServerBuildParameter) JobID {
	jobID := JobID(fmt.Sprintf("build/%s/%s/%s", param.NameSpace, param.ClusterName, param.ServerName))
	status := &JobStatus{
		ID:    jobID,
		Type:  JobTypeProvisioning,
		State: JobStatePending,
	}
	s.jobs.set(jobID, status)

	go func() {
		status.State = JobStateInFlight

		// build server
		builderClient := server.NewBuildersAPIClient(s.caller)
		builder := s.createBuilder(param)
		result, err := builder.Build(ctx, builderClient, zone)
		if err != nil {
			status.Error = err
			status.State = JobStateFailed
			return
		}
		if result != nil {
			status.Reference = &CloudObjectRef{ServerID: result.ServerID}
		}

		// build iso-image
		sv, err := s.serverOp().Read(ctx, zone, result.ServerID)
		if err != nil {
			status.Error = err
			status.State = JobStateFailed
			return
		}
		isoImage, err := s.buildISOImage(ctx, zone, sv, param)
		if err != nil {
			status.Error = err
			status.State = JobStateFailed
			return
		}
		status.Reference.ISOImageID = isoImage.ID

		// insert
		if err := s.serverOp().InsertCDROM(ctx, zone, sv.ID, &sacloud.InsertCDROMRequest{ID: isoImage.ID}); err != nil {
			status.Error = err
			status.State = JobStateFailed
			return
		}

		if err := s.serverOp().Boot(ctx, zone, sv.ID); err != nil {
			status.Error = err
			status.State = JobStateFailed
			return
		}

		_, err = sacloud.WaiterForUp(func() (state interface{}, err error) {
			return s.serverOp().Read(ctx, zone, sv.ID)
		}).WaitForState(ctx)
		if err != nil {
			status.Error = err
			status.State = JobStateFailed
			return
		}

		status.State = JobStateDone
	}()

	return jobID
}

func (s *serverClient) buildTagsFromContext(clusterName, nameSpace string, isControlPlane bool) sacloudtypes.Tags {
	return sacloudtypes.Tags{
		fmt.Sprintf("cluster=%s", clusterName),
		fmt.Sprintf("ns=%s", nameSpace),
		fmt.Sprintf("control-plane=%t", isControlPlane), // util.IsControlPlaneMachine(ctx.Machine)
	}
}

func (s *serverClient) createBuilder(param *ServerBuildParameter) *server.Builder {
	return &server.Builder{
		Name:            param.ServerName,
		CPU:             param.Spec.CPUs,
		MemoryGB:        param.Spec.MemoryGB,
		Commitment:      sacloudtypes.Commitments.Standard,
		Generation:      sacloudtypes.PlanGenerations.Default,
		InterfaceDriver: sacloudtypes.InterfaceDrivers.VirtIO,
		Description:     "", // TODO 何か入れる?
		Tags:            s.buildTagsFromContext(param.ClusterName, param.NameSpace, param.IsControlPlane),
		BootAfterCreate: false,                      // for insert ISO-Image with metadata
		NIC:             &server.SharedNICSetting{}, // TODO あとで直す
		DiskBuilders: []server.DiskBuilder{
			&server.FromDiskOrArchiveDiskBuilder{
				SourceArchiveID: sacloudtypes.StringID(param.SourceArchiveID),
				Name:            param.ServerName,
				SizeGB:          param.Spec.DiskGB,
				PlanID:          sacloudtypes.DiskPlans.SSD,
				Connection:      sacloudtypes.DiskConnections.VirtIO,
				Description:     "",
				Tags:            s.buildTagsFromContext(param.ClusterName, param.NameSpace, param.IsControlPlane),
			},
		},
	}
}

func (s *serverClient) buildISOImage(ctx context.Context, zone string, sv *sacloud.Server, param *ServerBuildParameter) (*sacloud.CDROM, error) {
	userData, err := s.generateUserData(param.BootstrapData)
	if err != nil {
		return nil, err
	}
	metaData, err := s.generateMetaData(sv)
	if err != nil {
		return nil, err
	}

	// create ISO image on SakuraCloud
	isoImage, ftpInfo, err := s.isoImageOp().Create(ctx, zone, &sacloud.CDROMCreateRequest{
		SizeMB:      5 * 1024,
		Name:        param.ServerName,
		Description: "",
		Tags:        sv.Tags,
	})
	if err != nil {
		return nil, err
	}

	// create tmp dir wor generate ISO image file
	tmpDir, err := ioutil.TempDir("", "caps-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir) // ignore error

	if err := ioutil.WriteFile(filepath.Join(tmpDir, "user-data"), userData, 0600); err != nil {
		return nil, err
	}
	if err := ioutil.WriteFile(filepath.Join(tmpDir, "meta-data"), metaData, 0600); err != nil {
		return nil, err
	}

	// generate cloud-init.iso
	if err := generateCISourceCmd(ctx, tmpDir).Run(); err != nil {
		return nil, err
	}

	isoFile, err := os.Open(filepath.Join(tmpDir, "cloud-init.iso"))
	if err != nil {
		return nil, err
	}
	defer isoFile.Close() // ignore error

	ftpsClient := ftps.NewClient(ftpInfo.User, ftpInfo.Password, ftpInfo.HostName)
	if err := ftpsClient.UploadFile("cloud-init.iso", isoFile); err != nil {
		return nil, err
	}

	// close FTP
	if err := s.isoImageOp().CloseFTP(ctx, zone, isoImage.ID); err != nil {
		return nil, err
	}

	return isoImage, nil
}

func (s *serverClient) generateUserData(bootstrapData string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(bootstrapData)
}

func (s *serverClient) generateMetaData(server *sacloud.Server) ([]byte, error) {
	mdTmpl := `{
  "instance-id": "{{.ID}}",
  "hostname": "{{.Name}}",
  "local-hostname": "{{.Name}}"
}`
	buf := bytes.NewBufferString("")
	tmpl, err := template.New("meta-data").Parse(mdTmpl)
	if err != nil {
		return nil, err
	}
	if err := tmpl.Execute(buf, server); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
