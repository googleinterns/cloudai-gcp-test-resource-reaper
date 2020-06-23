package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/reaper"
	"google.golang.org/api/option"
)

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

func (manager *ReaperManager) MonitorReapers(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-manager.quit:
			return
		case newReaper := <-manager.newReaper:
			manager.Reapers = append(manager.Reapers, newReaper)
			fmt.Println(manager.Reapers)
		default:
			for _, reaper := range manager.Reapers {
				reaper.RunOnSchedule(manager.ctx, manager.clientOptions...)
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
