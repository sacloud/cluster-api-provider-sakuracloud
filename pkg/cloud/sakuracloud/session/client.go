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
	"fmt"
	"net/http"
	"os"

	infrav1 "github.com/sacloud/cluster-api-provider-sakuracloud/api/v1alpha2"
	"github.com/sacloud/cluster-api-provider-sakuracloud/version"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

type Client struct {
	ServerAPI
	jobs *jobRegistry
}

func NewClient() *Client {
	ua := fmt.Sprintf("cluster-api-provider-sakuracloud/v%s (%s)", version.Version, infrav1.GroupVersion.String())

	caller := &sacloud.Client{
		AccessToken:            os.Getenv("SAKURACLOUD_ACCESS_TOKEN"),
		AccessTokenSecret:      os.Getenv("SAKURACLOUD_ACCESS_TOKEN_SECRET"),
		DefaultTimeoutDuration: sacloud.APIDefaultTimeoutDuration,
		UserAgent:              ua,
		AcceptLanguage:         sacloud.APIDefaultAcceptLanguage,
		RetryMax:               sacloud.APIDefaultRetryMax,
		RetryInterval:          sacloud.APIDefaultRetryInterval,
		HTTPClient:             http.DefaultClient,
	}

	caller.HTTPClient.Transport = &sacloud.RateLimitRoundTripper{
		Transport:       caller.HTTPClient.Transport,
		RateLimitPerSec: 3,
	}

	if os.Getenv("SAKURACLOUD_TRACE") != "" {
		caller.HTTPClient.Transport = &sacloud.TracingRoundTripper{
			Transport: caller.HTTPClient.Transport,
		}
	}

	jobs := &jobRegistry{}
	return &Client{
		ServerAPI: &serverClient{caller: caller, jobs: jobs},
		jobs:      jobs,
	}
}

func (c *Client) JobByID(id string) *JobStatus {
	return c.jobs.get(JobID(id))
}

func (c *Client) DeleteJob(id string) {
	c.jobs.delete(JobID(id))
}
