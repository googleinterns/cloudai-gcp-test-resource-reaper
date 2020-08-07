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
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/logger"
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

func init() {
	logger.CreateLogger()
}

type ReaperRunTestCase struct {
	Watchlist []*resources.WatchedResource
	Expected  *Reaper
}

var reaperRunTestCases = []ReaperRunTestCase{
	ReaperRunTestCase{
		[]*resources.WatchedResource{
			resources.NewWatchedResource(resources.NewResource("TestEarly", "testZone", earlyTime, reaperconfig.ResourceType_GCE_VM), "* 5 * * *"),
			resources.NewWatchedResource(resources.NewResource("TestFuture", "testZone", lateTime, reaperconfig.ResourceType_GCE_VM), "1 * * * *"),
			resources.NewWatchedResource(resources.NewResource("TestTwoMinuteAgo", "testZone", twoMinutesAgo, reaperconfig.ResourceType_GCE_VM), "59 * * * *"),
		},
		createTestReaper("testProject", "* * * * *", []*resources.WatchedResource{
			resources.NewWatchedResource(resources.NewResource("TestFuture", "testZone", lateTime, reaperconfig.ResourceType_GCE_VM), "1 * * * *"),
		}...),
	},
	ReaperRunTestCase{
		[]*resources.WatchedResource{
			resources.NewWatchedResource(resources.NewResource("TestTwoMinuteAgo", "testZone", twoMinutesAgo, reaperconfig.ResourceType_GCE_VM), "1 * * * *"),
		},
		createTestReaper("testProject", "* * * * *", []*resources.WatchedResource{
			resources.NewWatchedResource(resources.NewResource("TestTwoMinuteAgo", "testZone", twoMinutesAgo, reaperconfig.ResourceType_GCE_VM), "1 * * * *"),
		}...),
	},
	ReaperRunTestCase{
		[]*resources.WatchedResource{
			resources.NewWatchedResource(resources.NewResource("TestTwoMinuteAgo_1", "testZone", twoMinutesAgo, reaperconfig.ResourceType_GCE_VM), "59 * * * *"),
			resources.NewWatchedResource(resources.NewResource("TestTwoMinuteAgo_2", "testZone", twoMinutesAgo, reaperconfig.ResourceType_GCE_VM), "30 * * * *"),
		},
		createTestReaper("testProject", "* * * * *", []*resources.WatchedResource{
			resources.NewWatchedResource(resources.NewResource("TestTwoMinuteAgo_2", "testZone", twoMinutesAgo, reaperconfig.ResourceType_GCE_VM), "30 * * * *"),
		}...),
	},
}

func TestSweepThroughResources(t *testing.T) {
	server := createServer(deleteComputeEngineResourceHandler)
	defer server.Close()

	testClientOptions := getTestClientOptions(server)

	for _, testCase := range reaperRunTestCases {
		testReaper := createTestReaper("testProject", "* * * * *", testCase.Watchlist...)
		testReaper.FreezeTime(currentTime)

		testReaper.SweepThroughResources(testContext, testClientOptions...)
		if !areWatchlistsEqual(testReaper, testCase.Expected) {
			t.Errorf("Reaper not updated correctly after sweep through watched resources")
		}
	}
}

type UpdateReaperConfigTestCase struct {
	ReaperConfig *reaperconfig.ReaperConfig
	Expected     *Reaper
}

var updateReaperConfigTestCases = []UpdateReaperConfigTestCase{
	UpdateReaperConfigTestCase{
		createReaperConfig("SampleProject", "* * * * *"),
		createTestReaper("SampleProject", "* * * * *"),
	},
	UpdateReaperConfigTestCase{
		createReaperConfig("NewProjectID", "* * 10 * *"),
		createTestReaper("NewProjectID", "* * 10 * *"),
	},
	UpdateReaperConfigTestCase{
		createReaperConfig("AnotherProjectID", "59 23 31 12 7"),
		createTestReaper("AnotherProjectID", "59 23 31 12 7"),
	},
	UpdateReaperConfigTestCase{
		createReaperConfig("ProjectIDAgain", "@every 1h30m"),
		createTestReaper("ProjectIDAgain", "@every 1h30m"),
	},
}

func TestUpdateReaperConfig(t *testing.T) {
	testReaper := &Reaper{}
	for _, testCase := range updateReaperConfigTestCases {
		testReaper.UpdateReaperConfig(testCase.ReaperConfig)
		if strings.Compare(testReaper.ProjectID, testCase.Expected.ProjectID) != 0 {
			t.Errorf("Expected project id: %s, got: %s", testCase.Expected.ProjectID, testReaper.ProjectID)
		}
		if !reflect.DeepEqual(testReaper.Schedule, testCase.Expected.Schedule) {
			t.Error("Schedule not updated correctly")
		}
	}
}

type GetResourcesTestCase struct {
	ReaperConfig *reaperconfig.ReaperConfig
	Expected     *Reaper
}

var getResourcesTestCases = []GetResourcesTestCase{
	GetResourcesTestCase{
		createReaperConfig(
			"sampleProject", "* * * * *", createResourceConfig(reaperconfig.ResourceType_GCE_VM, "Test", "", "* * * * *", "testZone1"),
		),
		createTestReaper("sampleProject", "* * * * *", resources.CreateWatchlist(
			[]*resources.Resource{
				resources.NewResource("TestName", "testZone1", currentTime, reaperconfig.ResourceType_GCE_VM),
				resources.NewResource("TestingYetAnotherOne", "testZone1", currentTime, reaperconfig.ResourceType_GCE_VM),
			},
			"* * * * *",
		)...),
	},
	GetResourcesTestCase{
		createReaperConfig(
			"sampleProject", "* * * * *", createResourceConfig(reaperconfig.ResourceType_GCE_VM, "Test", "Another", "* * * * *", "testZone1"),
		),
		createTestReaper("sampleProject", "* * * * *", resources.CreateWatchlist(
			[]*resources.Resource{
				resources.NewResource("TestName", "testZone1", currentTime, reaperconfig.ResourceType_GCE_VM),
			},
			"* * * * *",
		)...),
	},
	GetResourcesTestCase{
		createReaperConfig(
			"sampleProject", "* * * * *", createResourceConfig(reaperconfig.ResourceType_GCE_VM, "Test", "", "* * * * *", "testZone1", "testZone2"),
		),
		createTestReaper("sampleProject", "* * * * *", resources.CreateWatchlist(
			[]*resources.Resource{
				resources.NewResource("TestName", "testZone1", currentTime, reaperconfig.ResourceType_GCE_VM),
				resources.NewResource("TestingYetAnotherOne", "testZone1", currentTime, reaperconfig.ResourceType_GCE_VM),
				resources.NewResource("TestThis", "testZone2", currentTime, reaperconfig.ResourceType_GCE_VM),
			},
			"* * * * *",
		)...),
	},
	GetResourcesTestCase{
		createReaperConfig(
			"sampleProject", "* * * * *", createResourceConfig(reaperconfig.ResourceType_GCE_VM, "Test", "Testing", "* * * * *", "testZone1", "testZone2"),
		),
		createTestReaper("sampleProject", "* * * * *", resources.CreateWatchlist(
			[]*resources.Resource{
				resources.NewResource("TestName", "testZone1", currentTime, reaperconfig.ResourceType_GCE_VM),
				resources.NewResource("TestThis", "testZone2", currentTime, reaperconfig.ResourceType_GCE_VM),
			},
			"* * * * *",
		)...),
	},
	GetResourcesTestCase{
		createReaperConfig(
			"sampleProject", "* * * * *", createResourceConfig(reaperconfig.ResourceType_GCE_VM, "Another", "", "* * * * *", "testZone1", "testZone2"),
		),
		createTestReaper("sampleProject", "* * * * *", resources.CreateWatchlist(
			[]*resources.Resource{
				resources.NewResource("AnotherName", "testZone1", currentTime, reaperconfig.ResourceType_GCE_VM),
				resources.NewResource("TestingYetAnotherOne", "testZone1", currentTime, reaperconfig.ResourceType_GCE_VM),
				resources.NewResource("IsThisAnotherName", "testZone2", currentTime, reaperconfig.ResourceType_GCE_VM),
			},
			"* * * * *",
		)...),
	},
}

func TestGetResources(t *testing.T) {
	server := createServer(getComputeEngineResourcesHandler)
	defer server.Close()

	testClientOptions := getTestClientOptions(server)
	testReaper := &Reaper{}

	setupTestData()
	for _, testCase := range getResourcesTestCases {
		testReaper.config = testCase.ReaperConfig
		testReaper.ProjectID = testCase.ReaperConfig.GetProjectId()

		testReaper.GetResources(testContext, testClientOptions...)
		if !areWatchlistsEqual(testReaper, testCase.Expected) {
			t.Errorf("GetResources did not get correct resources based off config")
		}
	}
}

type RunScheduleTestCase struct {
	Schedule string
	LastRun  time.Time
	Expected []*resources.Resource
}

var runScheduleTestCases = []RunScheduleTestCase{
	RunScheduleTestCase{"* * * * *", time.Time{}, []*resources.Resource{}},
	RunScheduleTestCase{"* * * 10 *", time.Time{}, []*resources.Resource{}},
	RunScheduleTestCase{"* 11 * * *", currentTime.Add(-1 * time.Hour), nil},
	RunScheduleTestCase{"* 10 * * *", currentTime.Add(-1 * time.Hour), []*resources.Resource{}},
	RunScheduleTestCase{"@every 1m", currentTime.Add(-2 * time.Minute), []*resources.Resource{}},
	RunScheduleTestCase{"@every 1h", currentTime.Add(-1 * time.Hour), []*resources.Resource{}},
}

func TestRunOnSchedule(t *testing.T) {
	for _, testCase := range runScheduleTestCases {
		reaper := createTestReaper("sampleProject", testCase.Schedule)
		reaper.FreezeClock(currentTime)
		reaper.lastRun = testCase.LastRun
		if result := reaper.RunOnSchedule(testContext); !reflect.DeepEqual(result, testCase.Expected) {
			t.Errorf("Reaper did run: %v, Should reaper run: %v", result != nil, testCase.Expected != nil)
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
func areWatchlistsEqual(result, expected *Reaper) bool {
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

func createReaperConfig(projectID, schedule string, resources ...*reaperconfig.ResourceConfig) *reaperconfig.ReaperConfig {
	return &reaperconfig.ReaperConfig{
		Resources: resources,
		Schedule:  schedule,
		ProjectId: projectID,
		Uuid:      "TestUUID",
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

func createTestReaper(projectID, schedule string, watchlist ...*resources.WatchedResource) *Reaper {
	parsedSchedule, _ := parseSchedule(schedule)
	return &Reaper{
		UUID:      "TestUUID",
		ProjectID: projectID,
		Watchlist: watchlist,
		Schedule:  parsedSchedule,
	}
}
