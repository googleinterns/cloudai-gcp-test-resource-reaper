package resources

import (
	"regexp"
	"time"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
)

type Resource struct {
	Name        string
	Zone        string
	TimeCreated time.Time
	Type        reaperconfig.ResourceType
}

func NewResource(name, zone string, timeCreated time.Time, resourceType reaperconfig.ResourceType) Resource {
	return Resource{name, zone, timeCreated, resourceType}
}

func (resource Resource) TimeAlive() float64 {
	timeAlive := time.Since(resource.TimeCreated)
	numSeconds := timeAlive.Seconds()
	return numSeconds
}

func FilterResources(resources []Resource, nameFilter, skipFilter string) []Resource {
	var filteredResources []Resource
	for _, resource := range resources {
		if ShouldAddResourceToWatchlist(resource, nameFilter, skipFilter) {
			filteredResources = append(filteredResources, resource)
		}
	}
	return filteredResources
}

// Empty skip filter doesn't catch anything,and name filter must be set
func ShouldAddResourceToWatchlist(resource Resource, nameFilter, skipFilter string) bool {
	if len(nameFilter) == 0 {
		return false
	}
	resourceName := resource.Name
	if len(skipFilter) > 0 {
		skipMatch, err := regexp.MatchString(skipFilter, resourceName)
		if err != nil || skipMatch {
			return false
		}
	}
	nameMatch, err := regexp.MatchString(nameFilter, resourceName)
	if err != nil {
		return false
	}
	return nameMatch
}
