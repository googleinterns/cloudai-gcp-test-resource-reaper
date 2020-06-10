package main

import (
	"time"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/resources"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
)

func main() {
	//cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow
	res := resources.NewResource("TestResource", "TestZone", time.Now(), reaperconfig.ResourceType_GCE_VM)
	watched := resources.NewWatchedResource(res, "1 * * * *")
	watched.ReadyForDeletion()
}
