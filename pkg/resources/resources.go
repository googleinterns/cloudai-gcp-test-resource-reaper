package resources

import (
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
	return nil
}

func ShouldAddResourceToWatchlist(resource Resource, nameFilter, skipFilter string) bool {
	return true
}
