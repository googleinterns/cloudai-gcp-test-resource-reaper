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
	"log"
	"strings"
	"sync"
	"time"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/reaper"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
	"google.golang.org/api/option"
)

// ReaperManager is a controller for all running reapers.
type ReaperManager struct {
	Reapers []*reaper.Reaper

	ctx           context.Context
	clientOptions []option.ClientOption
	newReaper     chan *reaper.Reaper
	deleteReaper  chan string
	quit          chan bool
}

// Start thread for each reaper
// https://golang.org/pkg/time/#Tick
// https://gist.github.com/ryanfitz/4191392

// NewReaperManager creates a new reaper manager.
func NewReaperManager(ctx context.Context, clientOptions ...option.ClientOption) *ReaperManager {
	return &ReaperManager{
		ctx:           ctx,
		clientOptions: clientOptions,
		newReaper:     make(chan *reaper.Reaper, 3),
		deleteReaper:  make(chan string, 3),
		quit:          make(chan bool, 1),
	}
}

func (manager *ReaperManager) Start() {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go manager.MonitorReapers(wg)
	wg.Wait()
}

// MonitorReapers is the controller for all running reapers. It continuously
// cycles between all running reapers to run a sweep. The method also checks
// whether a new reaper has been added to the manager, or if the the manager
// should be stopped. Note that MonitorReapers should be called in a separate
// goroutine.
func (manager *ReaperManager) MonitorReapers(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-manager.quit:
			log.Println("Quitting reaper manager")
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
		log.Printf("Added new reaper with UUID: %s", newReaper.UUID)
	case reaperUUID := <-manager.deleteReaper:
		deleteSuccess := manager.handleDeleteReaper(reaperUUID)
		if deleteSuccess {
			log.Printf("Reaper with UUID %s successfully deleted", reaperUUID)
		} else {
			log.Printf("Reaper with UUID %s unsuccessfully deleted", reaperUUID)
		}
	default:
		for _, reaper := range manager.Reapers {
			reaper.RunOnSchedule(manager.ctx, manager.clientOptions...)
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
	newReaper.UpdateReaperConfig(newReaperConfig)
	manager.newReaper <- newReaper
}

// Shutdown ends the reaper manager process.
func (manager *ReaperManager) Shutdown() {
	manager.quit <- true
}

// DeleteReaper sends a signal to delete a reaper with the given UUID.
func (manager *ReaperManager) DeleteReaper(uuid string) {
	manager.deleteReaper <- uuid
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
