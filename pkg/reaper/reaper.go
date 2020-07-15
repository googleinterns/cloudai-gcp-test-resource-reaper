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
	"time"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/clients"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/logger"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/resources"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
	"github.com/robfig/cron/v3"
	"google.golang.org/api/option"
)

// Reaper represents the resource reaper for a single GCP project. The reaper will
// run on a given schedule defined in cron time format.
type Reaper struct {
	UUID      string
	ProjectID string
	Watchlist []*resources.WatchedResource
	Schedule  cron.Schedule

	config  *reaperconfig.ReaperConfig
	lastRun time.Time
	*Clock
}

type Clock struct {
	instant time.Time
}

func (c *Clock) Now() time.Time {
	if c == nil {
		return time.Now()
	}
	return c.instant
}

func (reaper *Reaper) FreezeClock(instant time.Time) {
	if reaper.Clock == nil {
		reaper.Clock = &Clock{}
	}
	reaper.Clock.instant = instant
}

// NewReaper constructs a new reaper.
func NewReaper() *Reaper {
	return &Reaper{}
}

// RunOnSchedule updates the reaper's watchlist and runs a sweep if the current time is equal to or after
// the next schedule run time.
func (reaper *Reaper) RunOnSchedule(ctx context.Context, clientOptions ...option.ClientOption) bool {
	nextRun := reaper.Schedule.Next(reaper.lastRun)
	if reaper.lastRun.IsZero() || reaper.Clock.Now().After(nextRun) || reaper.Clock.Now().Equal(nextRun) {
		logger.Logf("Running reaper with UUID: %s\n", reaper.UUID)
		reaper.GetResources(ctx, clientOptions...)

		logger.Logf("Reaper %s sweeping through the following resources: ")
		reaper.SweepThroughResources(ctx, clientOptions...)
		reaper.lastRun = reaper.Clock.Now()
		return true
	}
	return false
}

// SweepThroughResources goes through all the resources in the reaper's Watchlist, and for each resource
// determines if it needs to be deleted. The necessary resources are deleted from GCP and the reaper's
// Watchlist is updated accordingly.
func (reaper *Reaper) SweepThroughResources(ctx context.Context, clientOptions ...option.ClientOption) {
	var updatedWatchlist []*resources.WatchedResource

	for _, watchedResource := range reaper.Watchlist {
		if watchedResource.IsReadyForDeletion() {
			resourceClient, err := getAuthedClient(ctx, reaper, watchedResource.Type, clientOptions...)
			if err != nil {
				logger.Error(err)
				continue
			}

			if err := resourceClient.DeleteResource(reaper.ProjectID, watchedResource.Resource); err != nil {
				deleteError := fmt.Errorf(
					"%s client failed to delete resource %s with the following error: %s",
					watchedResource.Type.String(), watchedResource.Name, err.Error(),
				)
				logger.Error(deleteError)
				continue
			}
			logger.Logf(
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
func (reaper *Reaper) UpdateReaperConfig(config *reaperconfig.ReaperConfig) error {
	reaper.config = config

	reaper.ProjectID = config.GetProjectId()
	reaper.UUID = config.GetUuid()

	parsedSchedule, err := parseSchedule(config.GetSchedule())
	reaper.Schedule = parsedSchedule
	return err
}

// GetResources gets all the GCP resources defined in the ReaperConfig, and adds them to the
// reaper's Watchlist. Note, if the same resource is referenced by multiple ResourceConfigs,
// then the TTL of that resource will be the one that deletes the resource the latest.
func (reaper *Reaper) GetResources(ctx context.Context, clientOptions ...option.ClientOption) {
	var newWatchlist []*resources.WatchedResource
	newWatchedResources := make(map[string]map[string]*resources.WatchedResource)

	resourceConfigs := reaper.config.GetResources()
	for _, resourceConfig := range resourceConfigs {
		resourceType := resourceConfig.GetResourceType()

		resourceClient, err := getAuthedClient(ctx, reaper, resourceType, clientOptions...)
		if err != nil {
			logger.Error(err)
			continue
		}

		filteredResources, err := resourceClient.GetResources(reaper.ProjectID, resourceConfig)
		if err != nil {
			getResourcesError := fmt.Errorf(
				"%s client failed to get resources with the following error: %s",
				resourceType.String(), err.Error(),
			)
			logger.Error(getResourcesError)
			continue
		}
		watchedResources := resources.CreateWatchlist(filteredResources, resourceConfig.GetTtl())

		// Check for duplicates. If one exists, update the TTL by the max
		for _, resource := range watchedResources {
			if _, isZoneWatched := newWatchedResources[resource.Zone]; !isZoneWatched {
				newWatchedResources[resource.Zone] = make(map[string]*resources.WatchedResource)
			}

			if _, alreadyWatched := newWatchedResources[resource.Zone][resource.Name]; alreadyWatched {
				newTTL, err := maxTTL(resource, newWatchedResources[resource.Zone][resource.Name])
				if err != nil {
					logger.Error(err)
					continue
				}
				newWatchedResources[resource.Zone][resource.Name].TTL = newTTL
			} else {
				newWatchedResources[resource.Zone][resource.Name] = resource
			}
		}
	}
	// Converting resources map into list
	for zone := range newWatchedResources {
		for _, resource := range newWatchedResources[zone] {
			newWatchlist = append(newWatchlist, resource)
		}
	}
	reaper.Watchlist = newWatchlist
}

// PrintWatchlist neatly prints the reaper's Watchlist.
func (reaper *Reaper) PrintWatchlist() {
	fmt.Print("Watchlist: ")
	for _, resource := range reaper.Watchlist {
		fmt.Printf("%s in %s, ", resource.Name, resource.Zone)
	}
	fmt.Print("\n")
}

// NewReaperConfig constructs a new ReaperConfig.
func NewReaperConfig(resources []*reaperconfig.ResourceConfig, schedule, projectID, uuid string) *reaperconfig.ReaperConfig {
	return &reaperconfig.ReaperConfig{
		Resources: resources,
		Schedule:  schedule,
		ProjectId: projectID,
		Uuid:      uuid,
	}
}

// NewResourceConfig constructs a new ResourceConfig.
func NewResourceConfig(resourceType reaperconfig.ResourceType, zones []string, nameFilter, skipFilter, ttl string) *reaperconfig.ResourceConfig {
	return &reaperconfig.ResourceConfig{
		ResourceType: resourceType,
		NameFilter:   nameFilter,
		SkipFilter:   skipFilter,
		Zones:        zones,
		Ttl:          ttl,
	}
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

// FreezeTime is a helper method for freezing the clocks of all resources in a reaper's
// Watchlist to a given instant.
func (reaper *Reaper) FreezeTime(instant time.Time) {
	for idx := range reaper.Watchlist {
		reaper.Watchlist[idx].FreezeClock(instant)
	}
}

// maxTTL is a helper function to determine which watched resource will be deleted later,
// and return its TTL.
func maxTTL(resourceA, resourceB *resources.WatchedResource) (string, error) {
	timeA, err := resourceA.GetDeletionTime()
	if err != nil {
		return "", fmt.Errorf("Parsing TTL failed with following error: %s", err.Error())
	}
	timeB, err := resourceB.GetDeletionTime()
	if err != nil {
		return "", fmt.Errorf("Parsing TTL failed with following error: %s", err.Error())
	}
	if timeA.After(timeB) {
		return resourceA.TTL, nil
	} else {
		return resourceB.TTL, nil
	}
}

// parseSchedule parses the cron time string that defined the reaper's
// run schedule, and either returns a Schedule struct, or nil if the
// schedule string is malformed.
func parseSchedule(schedule string) (cron.Schedule, error) {
	parsedSchedule, err := cron.ParseStandard(schedule)
	if err != nil {
		return nil, err
	}
	return parsedSchedule, nil
}
