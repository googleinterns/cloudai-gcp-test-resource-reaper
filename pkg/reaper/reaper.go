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

// Log to stack driver

// Optimization: Use DLL for watchlist, and have map value be pointer to element. O(1) deletion and search
type Reaper struct {
	UUID      string
	ProjectID string
	Watchlist []resources.WatchedResource
	Schedule  string

	// Helped structure for quickly determining if a resource is being watched.
	// { Zone : { Resource Name : Pointer to WatchedResource }}
	// watched map[string]map[string]*resources.WatchedResource
}

func NewReaper(ctx context.Context, config *reaperconfig.ReaperConfig, clientOptions ...option.ClientOption) *Reaper {
	reaper := &Reaper{}
	reaper.UpdateReaperConfig(ctx, config, clientOptions...)
	return reaper
}

func (reaper *Reaper) RunThroughResources(ctx context.Context, clientOptions ...option.ClientOption) {
	var updatedWatchlist []resources.WatchedResource

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
		} else {
			updatedWatchlist = append(updatedWatchlist, watchedResource)
		}
	}
	reaper.Watchlist = updatedWatchlist
}

func (reaper *Reaper) UpdateReaperConfig(ctx context.Context, config *reaperconfig.ReaperConfig, clientOptions ...option.ClientOption) {
	var newWatchlist []resources.WatchedResource

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

func (reaper *Reaper) freezeTime(instant time.Time) {
	for idx := range reaper.Watchlist {
		reaper.Watchlist[idx].FreezeClock(instant)
	}
}
