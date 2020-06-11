package reaper

import (
	"fmt"
	"log"
	"strings"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/clients"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/resources"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
	"google.golang.org/api/option"
)

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

type reaperStatus struct {
	AddedResources      []*resources.Resource
	DeletedResources    []*resources.Resource
	Failed              []error
	UnmodifiedResources []*resources.Resource
	UpdatedSchedule     bool
}

type reaperFailure struct {
	Resource *resources.Resource
	Error    error
}

func NewReaper(config *reaperconfig.ReaperConfig, clientOptions ...option.ClientOption) (Reaper, reaperStatus) {
	reaper := Reaper{}
	status := reaper.UpdateWatchedResources(config, clientOptions...)
	return reaper, status
}

func (reaper Reaper) RunThroughResources(clientOptions ...option.ClientOption) reaperStatus {
	status := reaperStatus{}

	for _, watchedResource := range reaper.Watchlist {
		if watchedResource.IsReadyForDeletion() {
			resourceClient, err := clients.GetClientForResource(watchedResource.Type)
			if err != nil {
				log.Fatal(err)
			}
			resourceClient.Auth(clientOptions...)
			err = resourceClient.DeleteResource(reaper.ProjectID, watchedResource.Resource)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	return status
}

func (reaper *Reaper) UpdateWatchedResources(config *reaperconfig.ReaperConfig, clientOptions ...option.ClientOption) reaperStatus {
	status := reaperStatus{}
	var newWatchlist []resources.WatchedResource

	resourceConfigs := config.GetResources()
	for _, resourceConfig := range resourceConfigs {
		resourceType := resourceConfig.GetResourceType()
		resourceClient, err := clients.GetClientForResource(resourceType)
		if err != nil {
			detailedErr := fmt.Errorf("%s client failed with the following error: %s", resourceType.String(), err.Error())
			status.logReaperFailure(detailedErr)
			continue
		}

		err = resourceClient.Auth(clientOptions...)
		if err != nil {
			detailedErr := fmt.Errorf("%s client failed authenticate with the following error: %s", resourceType.String(), err.Error())
			status.logReaperFailure(detailedErr)
			continue
		}

		filteredResources, err := resourceClient.GetResources(reaper.ProjectID, resourceConfig)
		if err != nil {
			detailedErr := fmt.Errorf("%s client failed to get resources with the following error: %s", resourceType.String(), err.Error())
			status.logReaperFailure(detailedErr)
			continue
		}
		watchedResources := resources.CreateWatchlist(filteredResources, resourceConfig.GetTtl())
		newWatchlist = append(newWatchlist, watchedResources...)
		status.logAddedResources(filteredResources)
	}
	reaper.Watchlist = newWatchlist

	if strings.Compare(config.GetSchedule(), reaper.Schedule) != 0 {
		reaper.Schedule = config.GetSchedule()
		status.UpdatedSchedule = true
	}
	return status
}

func (status *reaperStatus) logAddedResources(addedResources []resources.Resource) {
	for _, resource := range addedResources {
		status.AddedResources = append(status.AddedResources, &resource)
	}
}

func (status *reaperStatus) logDeletedResources(resource *resources.Resource) {
	status.DeletedResources = append(status.DeletedResources, resource)
}

func (status *reaperStatus) logReaperFailure(err error) {
	status.Failed = append(status.Failed, err)
}

func (status *reaperStatus) logUnmodifiedResources(unmodifiedResources []resources.Resource) {
	for _, resource := range unmodifiedResources {
		status.UnmodifiedResources = append(status.UnmodifiedResources, &resource)
	}
}
