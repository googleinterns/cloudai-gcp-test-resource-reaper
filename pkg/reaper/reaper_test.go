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

package reaper

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/resources"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
	"google.golang.org/api/option"
)

var (
	testContext    = context.Background()
	currentTime, _ = time.Parse("2006-01-02 15:04:05 -0700", "2020-06-17 10:00:00 -0400")

	earlyTime     = currentTime.AddDate(-10, 0, 0)
	lateTime      = currentTime.AddDate(10, 0, 0)
	twoMinutesAgo = currentTime.Add(-2 * time.Minute)

	// ProjectID -> ResourceType -> Zones -> Resource
	testData map[string]map[reaperconfig.ResourceType]map[string][]TestData
)

type ReaperRunTestCase struct {
	Watchlist []*resources.WatchedResource
	Expected  Reaper
}

var reaperRunTestCases = []ReaperRunTestCase{
	ReaperRunTestCase{
		[]*resources.WatchedResource{
			resources.NewWatchedResource(resources.NewResource("TestEarly", "testZone", earlyTime, reaperconfig.ResourceType_GCE_VM), "* 5 * * *"),
			resources.NewWatchedResource(resources.NewResource("TestFuture", "testZone", lateTime, reaperconfig.ResourceType_GCE_VM), "1 * * * *"),
			resources.NewWatchedResource(resources.NewResource("TestTwoMinuteAgo", "testZone", twoMinutesAgo, reaperconfig.ResourceType_GCE_VM), "59 * * * *"),
		},
		createTestReaper("testProject", []*resources.WatchedResource{
			resources.NewWatchedResource(resources.NewResource("TestFuture", "testZone", lateTime, reaperconfig.ResourceType_GCE_VM), "1 * * * *"),
		}...),
	},
	ReaperRunTestCase{
		[]*resources.WatchedResource{
			resources.NewWatchedResource(resources.NewResource("TestTwoMinuteAgo", "testZone", twoMinutesAgo, reaperconfig.ResourceType_GCE_VM), "1 * * * *"),
		},
		createTestReaper("testProject", []*resources.WatchedResource{
			resources.NewWatchedResource(resources.NewResource("TestTwoMinuteAgo", "testZone", twoMinutesAgo, reaperconfig.ResourceType_GCE_VM), "1 * * * *"),
		}...),
	},
	ReaperRunTestCase{
		[]*resources.WatchedResource{
			resources.NewWatchedResource(resources.NewResource("TestTwoMinuteAgo_1", "testZone", twoMinutesAgo, reaperconfig.ResourceType_GCE_VM), "59 * * * *"),
			resources.NewWatchedResource(resources.NewResource("TestTwoMinuteAgo_2", "testZone", twoMinutesAgo, reaperconfig.ResourceType_GCE_VM), "30 * * * *"),
		},
		createTestReaper("testProject", []*resources.WatchedResource{
			resources.NewWatchedResource(resources.NewResource("TestTwoMinuteAgo_2", "testZone", twoMinutesAgo, reaperconfig.ResourceType_GCE_VM), "30 * * * *"),
		}...),
	},
}

func TestRunThroughResources(t *testing.T) {
	server := createServer(deleteComputeEngineResourceHandler)
	defer server.Close()

	testClientOptions := getTestClientOptions(server)

	for _, testCase := range reaperRunTestCases {
		testReaper := createTestReaper("testProject", testCase.Watchlist...)
		testReaper.freezeTime(currentTime)

		testReaper.RunThroughResources(testContext, testClientOptions...)
		if !areWatchlistsEqual(testReaper, testCase.Expected) {
			t.Errorf("Reaper not updated correctly")
		}
	}
}

type UpdateReaperTestCase struct {
	ReaperConfig *reaperconfig.ReaperConfig
	Expected     Reaper
}

var updateReaperTestCases = []UpdateReaperTestCase{
	UpdateReaperTestCase{
		createReaperConfig(
			"sampleProject", "", createResourceConfig(reaperconfig.ResourceType_GCE_VM, "Test", "", "* * * * *", "testZone1"),
		),
		createTestReaper("sampleProject", resources.CreateWatchlist(
			[]*resources.Resource{
				resources.NewResource("TestName", "testZone1", currentTime, reaperconfig.ResourceType_GCE_VM),
				resources.NewResource("TestingYetAnotherOne", "testZone1", currentTime, reaperconfig.ResourceType_GCE_VM),
			},
			"* * * * *",
		)...),
	},
	UpdateReaperTestCase{
		createReaperConfig(
			"sampleProject", "", createResourceConfig(reaperconfig.ResourceType_GCE_VM, "Test", "Another", "* * * * *", "testZone1"),
		),
		createTestReaper("sampleProject", resources.CreateWatchlist(
			[]*resources.Resource{
				resources.NewResource("TestName", "testZone1", currentTime, reaperconfig.ResourceType_GCE_VM),
			},
			"* * * * *",
		)...),
	},
	UpdateReaperTestCase{
		createReaperConfig(
			"sampleProject", "", createResourceConfig(reaperconfig.ResourceType_GCE_VM, "Test", "", "* * * * *", "testZone1", "testZone2"),
		),
		createTestReaper("sampleProject", resources.CreateWatchlist(
			[]*resources.Resource{
				resources.NewResource("TestName", "testZone1", currentTime, reaperconfig.ResourceType_GCE_VM),
				resources.NewResource("TestingYetAnotherOne", "testZone1", currentTime, reaperconfig.ResourceType_GCE_VM),
				resources.NewResource("TestThis", "testZone2", currentTime, reaperconfig.ResourceType_GCE_VM),
			},
			"* * * * *",
		)...),
	},
	UpdateReaperTestCase{
		createReaperConfig(
			"sampleProject", "", createResourceConfig(reaperconfig.ResourceType_GCE_VM, "Test", "Testing", "* * * * *", "testZone1", "testZone2"),
		),
		createTestReaper("sampleProject", resources.CreateWatchlist(
			[]*resources.Resource{
				resources.NewResource("TestName", "testZone1", currentTime, reaperconfig.ResourceType_GCE_VM),
				resources.NewResource("TestThis", "testZone2", currentTime, reaperconfig.ResourceType_GCE_VM),
			},
			"* * * * *",
		)...),
	},
	UpdateReaperTestCase{
		createReaperConfig(
			"sampleProject", "", createResourceConfig(reaperconfig.ResourceType_GCE_VM, "Another", "", "* * * * *", "testZone1", "testZone2"),
		),
		createTestReaper("sampleProject", resources.CreateWatchlist(
			[]*resources.Resource{
				resources.NewResource("AnotherName", "testZone1", currentTime, reaperconfig.ResourceType_GCE_VM),
				resources.NewResource("TestingYetAnotherOne", "testZone1", currentTime, reaperconfig.ResourceType_GCE_VM),
				resources.NewResource("IsThisAnotherName", "testZone2", currentTime, reaperconfig.ResourceType_GCE_VM),
			},
			"* * * * *",
		)...),
	},
}

func TestUpdateReaperConfig(t *testing.T) {
	server := createServer(getComputeEngineResourcesHandler)
	defer server.Close()

	testClientOptions := getTestClientOptions(server)
	testReaper := Reaper{}

	setupTestData()
	for _, testCase := range updateReaperTestCases {
		testReaper.UpdateReaperConfig(testContext, testCase.ReaperConfig, testClientOptions...)
		if !areWatchlistsEqual(testReaper, testCase.Expected) {
			t.Errorf("Reaper not updated correctly")
		}
	}
}

func createServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handler))
}

func deleteComputeEngineResourceHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(`{"success": true}`))
}

type GetResourcesResponse struct {
	Items []TestData
}

func getComputeEngineResourcesHandler(w http.ResponseWriter, req *http.Request) {
	// Endpoint of the form: /{ProjectID}/zones/{ZoneName}/instances
	endpoint := req.URL.Path
	splitEndpoint := strings.Split(endpoint, "/")
	projectID := splitEndpoint[1]
	zone := splitEndpoint[3]

	res := GetResourcesResponse{testData[projectID][reaperconfig.ResourceType_GCE_VM][zone]}
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(res)
}

// Only checking if names are equal since test is setup to have unique names
func areWatchlistsEqual(result, expected Reaper) bool {
	if len(result.Watchlist) != len(expected.Watchlist) {
		return false
	}
	var resultResourceNames = make(map[string]bool)
	for _, resource := range result.Watchlist {
		resultResourceNames[resource.Name] = true
	}
	for _, resource := range expected.Watchlist {
		if _, exists := resultResourceNames[resource.Name]; !exists {
			return false
		}
		delete(resultResourceNames, resource.Name)
	}
	return true
}

func getTestClientOptions(server *httptest.Server) []option.ClientOption {
	return []option.ClientOption{
		option.WithHTTPClient(server.Client()),
		option.WithEndpoint(server.URL),
	}
}

type TestData struct {
	Name              string
	CreationTimestamp string
}

func setupTestData() {
	testData = map[string]map[reaperconfig.ResourceType]map[string][]TestData{
		"sampleProject": {
			reaperconfig.ResourceType_GCE_VM: {
				"testZone1": []TestData{
					TestData{"TestName", currentTime.Format(time.RFC3339)},
					TestData{"AnotherName", currentTime.Format(time.RFC3339)},
					TestData{"TestingYetAnotherOne", currentTime.Format(time.RFC3339)},
				},
				"testZone2": []TestData{
					TestData{"TestThis", currentTime.Format(time.RFC3339)},
					TestData{"IsThisAnotherName", currentTime.Format(time.RFC3339)},
				},
			},
		},
	}
}

func createReaperConfig(projectID, skipFilter string, resources ...*reaperconfig.ResourceConfig) *reaperconfig.ReaperConfig {
	return &reaperconfig.ReaperConfig{
		Resources:  resources,
		Schedule:   "TestSchedule",
		SkipFilter: skipFilter,
		ProjectId:  projectID,
		Uuid:       "TestUUID",
	}
}

func createResourceConfig(resourceType reaperconfig.ResourceType, nameFilter, skipFilter, ttl string, zones ...string) *reaperconfig.ResourceConfig {
	return &reaperconfig.ResourceConfig{
		ResourceType: resourceType,
		NameFilter:   nameFilter,
		SkipFilter:   skipFilter,
		Zones:        zones,
		Ttl:          ttl,
	}
}

func createTestReaper(projectID string, watchlist ...*resources.WatchedResource) Reaper {
	return Reaper{
		UUID:      "TestUUID",
		ProjectID: projectID,
		Watchlist: watchlist,
		Schedule:  "TestSchedule",
	}
}
