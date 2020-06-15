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
	"context"
	"time"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
	gce "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/resources"
)

// Auth authenticates the client to access Compute Engine resources. See
// https://pkg.go.dev/google.golang.org/api/option?tab=doc for more
// information about passing options.
func (client *GCEClient) Auth(ctx context.Context, opts ...option.ClientOption) error {
	authedClient, err := gce.NewService(ctx, opts...)
	if err != nil {
		return err
	}
	client.Client = authedClient
	return nil
}

// GetResources gets the Compute Engine instances that pass the filters defined in the ResourceConfig
func (client *GCEClient) GetResources(projectID string, config *reaperconfig.ResourceConfig) ([]resources.Resource, error) {
	var instances []resources.Resource
	zones := config.GetZones()
	for _, zone := range zones {
		zoneInstancesCall := client.Client.Instances.List(projectID, zone)
		// Info on filtering: https://cloud.google.com/compute/docs/reference/rest/v1/instances/list
		// zoneInstancesCall.Filter()
		instancesInZone, err := zoneInstancesCall.Do()
		if err != nil {
			return nil, err
		}
		for _, instance := range instancesInZone.Items {
			timeCreated, _ := time.Parse(time.RFC3339, instance.CreationTimestamp)
			parsedResource := resources.NewResource(instance.Name, zone, timeCreated, reaperconfig.ResourceType_GCE_VM)
			if resources.ShouldAddResourceToWatchlist(parsedResource, config.GetNameFilter(), config.GetSkipFilter()) {
				instances = append(instances, parsedResource)
			}
		}
	}
	return instances, nil
}

// DeleteResource deletes the specificed Compute Engine instance.
func (client *GCEClient) DeleteResource(projectID string, resource resources.Resource) error {
	deleteInstanceCall := client.Client.Instances.Delete(projectID, resource.Zone, resource.Name)
	_, err := deleteInstanceCall.Do()
	return err
}
