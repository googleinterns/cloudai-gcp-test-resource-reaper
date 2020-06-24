package integration_tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/manager"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/reaper"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
)

func TestReaperManagerIntegration(t *testing.T) {
	reaperManager := manager.NewReaperManager(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go reaperManager.MonitorReapers(wg)

	resources := []*reaperconfig.ResourceConfig{
		reaper.NewResourceConfig(reaperconfig.ResourceType_GCE_VM, []string{"us-east1-b", "us-east1-c"}, "test", "skip", "9 7 * * *"),
		reaper.NewResourceConfig(reaperconfig.ResourceType_GCE_VM, []string{"us-east1-b"}, "another", "", "1 * * * *"),
		reaper.NewResourceConfig(reaperconfig.ResourceType_GCE_VM, []string{"us-east1-b"}, "another-resource-1", "", "* * * 10 *"),
	}
	reaperConfig := reaper.NewReaperConfig(resources, "@every 1m", "SkipFilter", projectID, "TestUUID")
	newReaper := reaper.NewReaper()
	newReaper.UpdateReaperConfig(context.Background(), reaperConfig)
	newReaper.FreezeTime(time.Now().AddDate(0, 1, 0))

	reaperManager.AddReaper(newReaper)

	wg.Wait()
}
