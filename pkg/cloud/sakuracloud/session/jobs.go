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
	"sync"

	sacloudtypes "github.com/sacloud/libsacloud/v2/sacloud/types"
)

type JobID string

type JobStatus struct {
	ID        JobID
	Type      JobType
	State     JobState
	Reference *CloudObjectRef
	Error     error
}

type JobType string

const (
	JobTypePending      JobState = ""
	JobTypeProvisioning          = "provisioning"
	JobTypeCleaning              = "cleaning"
)

type JobState string

const (
	JobStatePending  JobState = ""
	JobStateInFlight          = "inflight"
	JobStateDone              = "done"
	JobStateFailed            = "failed"
)

type CloudObjectRef struct {
	ServerID   sacloudtypes.ID
	ISOImageID sacloudtypes.ID
}

type jobRegistry struct {
	jobs sync.Map
}

func (j *jobRegistry) get(id JobID) *JobStatus {
	status, ok := j.jobs.Load(id)
	if ok {
		return status.(*JobStatus)
	}
	return nil
}

func (j *jobRegistry) set(id JobID, status *JobStatus) {
	j.jobs.Store(id, status)
}

func (j *jobRegistry) delete(id JobID) {
	j.jobs.Delete(id)
}
