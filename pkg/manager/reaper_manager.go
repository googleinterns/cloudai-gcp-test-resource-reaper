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

package manager

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/logger"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/reaper"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/resources"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
	"google.golang.org/api/option"
)

// ReaperManager is a controller for all running reapers.
type ReaperManager struct {
	Reapers []*reaper.Reaper

	ctx              context.Context
	clientOptions    []option.ClientOption
	newReaper        chan *reaper.Reaper
	deleteReaper     chan string
	updateReaper     chan *reaperconfig.ReaperConfig
	quit             chan bool
	deletedResources []*resources.Resource
}

// NewReaperManager creates a new reaper manager.
func NewReaperManager(ctx context.Context, clientOptions ...option.ClientOption) *ReaperManager {
	return &ReaperManager{
		ctx:           ctx,
		clientOptions: clientOptions,
		newReaper:     make(chan *reaper.Reaper, 3),
		deleteReaper:  make(chan string, 3),
		updateReaper:  make(chan *reaperconfig.ReaperConfig, 3),
		quit:          make(chan bool, 1),
	}
}

// MonitorReapers is the controller for all running reapers. It continuously
// cycles between all running reapers to run a sweep. The method also checks
// whether a new reaper has been added to the manager, or if the the manager
// should be stopped. Note that MonitorReapers should be called in a separate
// goroutine.
func (manager *ReaperManager) MonitorReapers() {
	logger.Log("Starting Reaper Manager")
	for {
		select {
		case <-manager.quit:
			logger.Log("Quitting reaper manager")
			return
		default:
			manager.sweepReapers()
		}
		time.Sleep(time.Second)
	}
}

// sweep handles the logic for a single sweep of all the reapers monitored
// by the manager. Note that this does not handle top-levek manager operations
// such as a shutdown.
func (manager *ReaperManager) sweepReapers() {
	select {
	case newReaper := <-manager.newReaper:
		manager.Reapers = append(manager.Reapers, newReaper)
		logger.Logf("Added new reaper with UUID: %s", newReaper.UUID)
	case reaperUUID := <-manager.deleteReaper:
		deleteSuccess := manager.handleDeleteReaper(reaperUUID)
		if deleteSuccess {
			logger.Logf("Reaper with UUID %s successfully deleted", reaperUUID)
		} else {
			logger.Logf("Reaper with UUID %s does not exist", reaperUUID)
		}
	case newReaperConfig := <-manager.updateReaper:
		err := manager.handleUpdateReaper(newReaperConfig)
		if err != nil {
			logger.Error(err)
		}
		logger.Logf("Reaper with UUID %s successfully updated", newReaperConfig.Uuid)
	default:
		for _, reaper := range manager.Reapers {
			deletedResources := reaper.RunOnSchedule(manager.ctx, manager.clientOptions...)
			if deletedResources != nil && len(deletedResources) > 0 {
				manager.deletedResources = append(manager.deletedResources, deletedResources...)
				rep, _ := manager.Report()
				fmt.Println(rep)
			}
		}
	}
}

// AddReaper adds a reaper to the manager.
func (manager *ReaperManager) AddReaper(newReaper *reaper.Reaper) {
	manager.newReaper <- newReaper
}

// AddReaperFromConfig adds a reaper to the manager from a ReaperConfig.
func (manager *ReaperManager) AddReaperFromConfig(newReaperConfig *reaperconfig.ReaperConfig) {
	newReaper := reaper.NewReaper()
	err := newReaper.UpdateReaperConfig(newReaperConfig)
	if err != nil {
		logger.Error(fmt.Errorf("error adding reaper: %v", err))
		return
	}
	manager.newReaper <- newReaper
}

// DeleteReaper sends a signal to delete a reaper with the given UUID.
func (manager *ReaperManager) DeleteReaper(uuid string) {
	manager.deleteReaper <- uuid
}

// UpdateReaper sends a signal to update a reaper with UUID given in the config.
func (manager *ReaperManager) UpdateReaper(config *reaperconfig.ReaperConfig) {
	manager.updateReaper <- config
}

// Shutdown ends the reaper manager process.
func (manager *ReaperManager) Shutdown() {
	manager.quit <- true
}

// handleDeleteReaper deletes the reaper with the given UUID, and returns whether the delete
// was successful. Note that false is returned if no reaper exists with the given UUID.
func (manager *ReaperManager) handleDeleteReaper(uuid string) bool {
	for idx, watchedReaper := range manager.Reapers {
		if strings.Compare(watchedReaper.UUID, uuid) == 0 {
			watchedReaper = nil
			manager.Reapers = append(manager.Reapers[:idx], manager.Reapers[idx+1:]...)
			return true
		}
	}
	return false
}

// handleUpdateReaper updates the reaper with the given UUID from the config, and returns
// whether the update was successful.
func (manager *ReaperManager) handleUpdateReaper(config *reaperconfig.ReaperConfig) error {
	for _, watchedReaper := range manager.Reapers {
		if strings.Compare(watchedReaper.UUID, config.GetUuid()) == 0 {
			err := watchedReaper.UpdateReaperConfig(config)
			return err
		}
	}
	return fmt.Errorf("Reaper with UUID %s does not exist", config.GetUuid())
}

// ListReapers returns a list of reapers being managed by the ReaperManager.
func (manager *ReaperManager) ListReapers() []*reaper.Reaper {
	return manager.Reapers
}

// GetReaper returns the reaper with the given UUID, or returns nil
// if no such reaper exists.
func (manager *ReaperManager) GetReaper(uuid string) *reaper.Reaper {
	for _, watchedReaper := range manager.Reapers {
		if strings.Compare(watchedReaper.UUID, uuid) == 0 {
			return watchedReaper
		}
	}
	return nil
}

func (manager *ReaperManager) Report() (string, error) {
	var report strings.Builder

	report.WriteString("Reaper Manager Report\n\nRunning Reapers:\n")
	for _, reaper := range manager.Reapers {
		report.WriteString(fmt.Sprintf("\u2022 %s\n", reaper.UUID))
	}

	report.WriteString("\nWatched Resources:\n")
	for _, reaper := range manager.Reapers {
		for _, watchedResource := range reaper.Watchlist {
			deletionTime, err := watchedResource.GetDeletionTime()
			if err != nil {
				return "", err
			}
			report.WriteString(
				fmt.Sprintf(
					"\u2022 %s in %s to be deleted at %s\n",
					watchedResource.Name,
					watchedResource.Zone,
					deletionTime.Format("2006-01-02 15:04:05"),
				),
			)
		}
	}

	report.WriteString("\nDeleted Resources:\n")
	for _, deletedResource := range manager.deletedResources {
		report.WriteString(fmt.Sprintf("\u2022 %s in %s\n", deletedResource.Name, deletedResource.Zone))
	}
	return report.String(), nil
}
