package main

import (
	"context"
	"sync"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/manager"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/reaper"
)

func main() {
	reaperManager := manager.NewReaperManager(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go reaperManager.MonitorReapers(wg)
	config := reaper.NewReaperConfig(
		
	)
	wg.Wait()
}
