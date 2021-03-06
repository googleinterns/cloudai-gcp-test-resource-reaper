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

package resources

import (
	"regexp"
	"time"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
	"github.com/robfig/cron/v3"
)

// A Resource represents a single GCP resource instance of any
// type supported by the Reaper.
type Resource struct {
	Name        string
	Zone        string
	TimeCreated time.Time
	Type        reaperconfig.ResourceType
}

// NewResource constructs a Resource struct.
func NewResource(name, zone string, timeCreated time.Time, resourceType reaperconfig.ResourceType) *Resource {
	return &Resource{name, zone, timeCreated, resourceType}
}

// TimeAlive returns how long a resource has been running.
func (resource *Resource) TimeAlive() float64 {
	timeAlive := time.Since(resource.TimeCreated)
	numSeconds := timeAlive.Seconds()
	return numSeconds
}

// Clock is a mock struct for handling time dependency for tests.
type Clock struct {
	instant time.Time
}

// Now returns the either time frozen time that was set for testing
// purposes, or the current time.
func (c *Clock) Now() time.Time {
	if c == nil {
		return time.Now()
	}
	return c.instant
}

// WatchedResource represents a resource that the Reaper is monitoring.
type WatchedResource struct {
	*Resource
	TTL   string
	clock *Clock
}

// NewWatchedResource constructs a WatchedResource.
func NewWatchedResource(resource *Resource, ttl string) *WatchedResource {
	return &WatchedResource{Resource: resource, TTL: ttl}
}

// FreezeClock sets the clock's current time to instant. This is to be
// used during testing.
func (resource *WatchedResource) FreezeClock(instant time.Time) {
	if resource.clock == nil {
		resource.clock = &Clock{}
	}
	resource.clock.instant = instant
}

// IsReadyForDeletion returns if a WatchedResource is past its time to live (TTL)
// based of the current time of the Clock.
func (resource *WatchedResource) IsReadyForDeletion() bool {
	deletionTime, err := resource.GetDeletionTime()
	if err != nil {
		return false
	}
	return resource.clock.Now().After(deletionTime)
}

func (resource *WatchedResource) GetDeletionTime() (time.Time, error) {
	// Using Cron time format doesn't give a duration, but instead a format of what the time should
	// look like when deleting
	schedule, err := cron.ParseStandard(resource.TTL)
	if err != nil {
		return time.Time{}, err
	}
	return schedule.Next(resource.TimeCreated), nil

}

// ShouldAddResourceToWatchlist determines whether a Resource should be watched
// by checking if its name matches the skip filter or name filter regex from the
// ResourceConfig and ReaperConfig. If a resource matches both the skip filter
// and name filter, then the skip filter wins and the resource will NOT be watched.
// An empty string for the skip filter will be interpreted as unset, and therefore
// will not match any resources.
func ShouldAddResourceToWatchlist(resource *Resource, nameFilter, skipFilter string) bool {
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

// CreateWatchlist creates a list of WatchedResources with a given time to
// live (TTL).
func CreateWatchlist(resources []*Resource, ttl string) []*WatchedResource {
	var watchlist []*WatchedResource
	for _, resource := range resources {
		watchedResource := NewWatchedResource(resource, ttl)
		watchlist = append(watchlist, watchedResource)
	}
	return watchlist
}
