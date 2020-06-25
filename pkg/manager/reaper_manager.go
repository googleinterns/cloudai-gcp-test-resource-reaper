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
		newReaper:     make(chan *reaper.Reaper),
		quit:          make(chan bool),
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
		case newReaper := <-manager.newReaper:
			manager.Reapers = append(manager.Reapers, newReaper)
			log.Printf("Added new reaper with UUID: %s", newReaper.UUID)
		default:
			for _, reaper := range manager.Reapers {
				reaper.RunOnSchedule(manager.ctx, manager.clientOptions...)
			}
		}
		time.Sleep(time.Second)
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

// Quit ends the reaper manager process.
func (manager *ReaperManager) Quit() {
	manager.quit <- true
}
