package manager

import (
	"fmt"
	"time"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/reaper"
)

type ReaperManager struct {
	Reapers   []*reaper.Reaper
	newReaper chan *reaper.Reaper
	quit      chan bool
}

// Start thread for each reaper
// https://golang.org/pkg/time/#Tick
// https://gist.github.com/ryanfitz/4191392

func NewReaperManager() *ReaperManager {
	return &ReaperManager{
		newReaper: make(chan *reaper.Reaper),
		quit:      make(chan bool),
	}
}

func (manager *ReaperManager) Start() {
	go manager.MonitorReapers()
}

func (manager *ReaperManager) MonitorReapers() {
	for {
		select {
		case <-manager.quit:
			return
		case newReaper := <-manager.newReaper:
			manager.Reapers = append(manager.Reapers, newReaper)
			fmt.Println(manager.Reapers)
		default:
			for _, reaper := range manager.Reapers {
				reaper.RunOnSchedule()
			}
		}
		time.Sleep(time.Second)
	}
}

func (manager *ReaperManager) AddReaper(newReaper *reaper.Reaper) {
	manager.newReaper <- newReaper
}

func (manager *ReaperManager) Quit() {
	manager.quit <- true
}

// infinite loop - run through all reapers and call their run on scehedule emthod
