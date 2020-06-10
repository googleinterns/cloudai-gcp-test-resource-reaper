package reaper

import (
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/resources"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
)

type Reaper struct {
	Watchlist []resources.WatchedResources
}

func (reaper Reaper) RunThroughResources() {

}

func (reaper Reaper) UpdateWatchedResources(config reaperconfig.ReaperConfig) {

}
