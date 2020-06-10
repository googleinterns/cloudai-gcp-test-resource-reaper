// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package clients

import (
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/resources"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
	gce "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// A Client represents the interface between the Reaper and GCP. Each resource
// supported by the reaper will have its own client that implements the following
// interface:
//  - Auth authenticates the client. Passing options changes how authentication
//    occurs. See https://pkg.go.dev/google.golang.org/api/option?tab=doc for
//    more details.
//  - GetResources returns a list of Resources that are match the ResourceConfig.
//  - DeleteResource deletes the specified resource.
type Client interface {
	Auth(opts ...option.ClientOption) error
	GetResources(projectID string, config *reaperconfig.ResourceConfig) ([]resources.Resource, error)
	DeleteResource(projectID string, resource resources.Resource) error
}

// Client for a Compute Engine Resource.
type GCEClient struct {
	Client *gce.Service
}
