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

package reaper

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/clients"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/resources"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
	"google.golang.org/api/option"
)

// Reaper represents the resource reaper for a single GCP project. The reaper will
// run on a given schedule defined in cron time format.
type Reaper struct {
	UUID      string
	ProjectID string
	Watchlist []*resources.WatchedResource
	Schedule  string
}

// NewReaper constructs a new reaper.
func NewReaper(ctx context.Context, config *reaperconfig.ReaperConfig, clientOptions ...option.ClientOption) *Reaper {
	reaper := &Reaper{}
	reaper.UpdateReaperConfig(ctx, config, clientOptions...)
	return reaper
}

// RunThroughResources goes through all the resources in the reaper's Watchlist, and for each resource
// determines if it needs to be deleted. The necessary resources are deleted from GCP and the reaper's
// Watchlist is updated accordingly.
func (reaper *Reaper) RunThroughResources(ctx context.Context, clientOptions ...option.ClientOption) {
	var updatedWatchlist []*resources.WatchedResource

	for _, watchedResource := range reaper.Watchlist {
		if watchedResource.IsReadyForDeletion() {
			resourceClient, err := getAuthedClient(ctx, reaper, watchedResource.Type, clientOptions...)
			if err != nil {
				log.Println(err)
				continue
			}

			if err := resourceClient.DeleteResource(reaper.ProjectID, watchedResource.Resource); err != nil {
				deleteError := fmt.Errorf(
					"%s client failed to delete resource %s with the following error: %s",
					watchedResource.Type.String(), watchedResource.Name, err.Error(),
				)
				log.Println(deleteError)
				continue
			}
			log.Printf(
				"Deleted %s resource %s in zone %s\n",
				watchedResource.Type.String(), watchedResource.Name, watchedResource.Zone,
			)
		} else {
			updatedWatchlist = append(updatedWatchlist, watchedResource)
		}
	}
	reaper.Watchlist = updatedWatchlist
}

// UpdateReaperConfig updates the reaper from a given ReaperConfig proto.
func (reaper *Reaper) UpdateReaperConfig(ctx context.Context, config *reaperconfig.ReaperConfig, clientOptions ...option.ClientOption) {
	var newWatchlist []*resources.WatchedResource

	if len(config.GetProjectId()) > 0 {
		reaper.ProjectID = config.GetProjectId()
	}
	if len(config.GetUuid()) > 0 {
		reaper.UUID = config.GetUuid()
	}
	if len(config.GetSchedule()) > 0 {
		reaper.Schedule = config.GetSchedule()
	}

	resourceConfigs := config.GetResources()
	for _, resourceConfig := range resourceConfigs {
		resourceType := resourceConfig.GetResourceType()

		resourceClient, err := getAuthedClient(ctx, reaper, resourceType, clientOptions...)
		if err != nil {
			log.Println(err)
			continue
		}

		filteredResources, err := resourceClient.GetResources(reaper.ProjectID, resourceConfig)
		if err != nil {
			getResourcesError := fmt.Errorf(
				"%s client failed to get resources with the following error: %s",
				resourceType.String(), err.Error(),
			)
			log.Println(getResourcesError)
			continue
		}
		watchedResources := resources.CreateWatchlist(filteredResources, resourceConfig.GetTtl())
		newWatchlist = append(newWatchlist, watchedResources...)
	}
	reaper.Watchlist = newWatchlist
}

// getAuthedClient is a helper method for getting an authenticated GCP client for a given resource type.
func getAuthedClient(ctx context.Context, reaper *Reaper, resourceType reaperconfig.ResourceType, clientOptions ...option.ClientOption) (clients.Client, error) {
	resourceClient, err := clients.NewClient(resourceType)
	if err != nil {
		clientError := fmt.Errorf(
			"%s client failed with the following error: %s",
			resourceType.String(), err.Error(),
		)
		return nil, clientError
	}

	err = resourceClient.Auth(ctx, clientOptions...)
	if err != nil {
		authError := fmt.Errorf(
			"%s client failed authenticate with the following error: %s",
			resourceType.String(), err.Error(),
		)
		return nil, authError
	}

	return resourceClient, nil
}

// freezeTime is a helper method for freezing the clocks of all resources in a reaper's
// Watchlist to a given instant.
func (reaper *Reaper) freezeTime(instant time.Time) {
	for idx := range reaper.Watchlist {
		reaper.Watchlist[idx].FreezeClock(instant)
	}
}
