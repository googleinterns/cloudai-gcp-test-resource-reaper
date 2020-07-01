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
	if testing.Short() {
		t.Skip("Skipping reaper integration test in short mode")
	}

	projectID, _, err := ReadConfigFile()
	if err != nil {
		t.Error(err.Error())
	}

	reaperManager := manager.NewReaperManager(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go reaperManager.MonitorReapers(wg)

	resources := []*reaperconfig.ResourceConfig{
		reaper.NewResourceConfig(reaperconfig.ResourceType_GCE_VM, []string{"us-east1-b", "us-east1-c"}, "test", "skip", "@every 1m"),
		reaper.NewResourceConfig(reaperconfig.ResourceType_GCE_VM, []string{"us-east1-b"}, "another", "", "1 * * * *"),
		reaper.NewResourceConfig(reaperconfig.ResourceType_GCE_VM, []string{"us-east1-b"}, "another-resource-1", "", "* * * 10 *"),
	}
	reaperConfig := reaper.NewReaperConfig(resources, "@every 1m", projectID, "TestUUID")

	newReaper := reaper.NewReaper()
	newReaper.UpdateReaperConfig(reaperConfig)
	newReaper.GetResources(context.Background())
	newReaper.FreezeTime(time.Now().AddDate(0, 1, 0))

	reaperManager.AddReaper(newReaper)

	wg.Wait()
}
