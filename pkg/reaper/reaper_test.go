package reaper_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/reaper"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/resources"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
	"google.golang.org/api/option"
)

// Use 2/3 types of resources and their clients for testing
var (
	testContext = context.Background()

	earlyTime   = time.Now().AddDate(-10, 0, 0)
	lateTime    = time.Now().AddDate(10, 0, 0)
	minuteAgo   = time.Now().Add(-1 * time.Minute)
	minuteLater = time.Now().Add(time.Minute)

	// ProjectID -> ResourceType -> Zones -> Resource
	testData map[string]map[reaperconfig.ResourceType]map[string][]TestData
)

type ReaperRunTestCase struct {
	Watchlist []resources.WatchedResource
	Expected  reaper.Reaper
}

var reaperRunTestCases = []ReaperRunTestCase{
	ReaperRunTestCase{
		[]resources.WatchedResource{
			resources.NewWatchedResource(resources.NewResource("Name", "Zone", earlyTime, reaperconfig.ResourceType_GCE_VM), "* * * * *"),
		},
		createTestReaper("project1", resources.CreateWatchlist(
			[]resources.Resource{
				resources.NewResource("Name", "Zone", earlyTime, reaperconfig.ResourceType_GCE_VM),
			},
			"* * * * *",
		)...),
	},
}

func TestRunThroughResources(t *testing.T) {
	server := createServer(deleteComputeEngineResourceHandler)
	defer server.Close()

	testClientOptions := getTestClientOptions(server)

	for _, testCase := range reaperRunTestCases {
		testReaper := createTestReaper("project1", testCase.Watchlist...)
		testReaper.RunThroughResources(testContext, testClientOptions...)
		if !areWatchlistsEqual(testReaper, testCase.Expected) {
			fmt.Println("FAIL")
			t.Errorf("Reaper not updated correctly")
		}
	}
}

type UpdateReaperTestCase struct {
	ReaperConfig *reaperconfig.ReaperConfig
	Expected     reaper.Reaper
}

var updateReaperTestCases = []UpdateReaperTestCase{
	UpdateReaperTestCase{
		createReaperConfig(
			"project2", "", createResourceConfig(reaperconfig.ResourceType_GCE_VM, "Test", "", "* * * * *", "testZone1"),
		),
		createTestReaper("project2", resources.CreateWatchlist(
			[]resources.Resource{
				resources.NewResource("TestName", "testZone1", earlyTime, reaperconfig.ResourceType_GCE_VM),
				resources.NewResource("TestingYetAnotherOne", "testZone1", earlyTime, reaperconfig.ResourceType_GCE_VM),
			},
			"* * * * *",
		)...),
	},
	UpdateReaperTestCase{
		createReaperConfig(
			"project2", "", createResourceConfig(reaperconfig.ResourceType_GCE_VM, "Test", "Another", "* * * * *", "testZone1"),
		),
		createTestReaper("project2", resources.CreateWatchlist(
			[]resources.Resource{
				resources.NewResource("TestName", "testZone1", earlyTime, reaperconfig.ResourceType_GCE_VM),
			},
			"* * * * *",
		)...),
	},
	UpdateReaperTestCase{
		createReaperConfig(
			"project2", "", createResourceConfig(reaperconfig.ResourceType_GCE_VM, "Test", "", "* * * * *", "testZone1", "testZone2"),
		),
		createTestReaper("project2", resources.CreateWatchlist(
			[]resources.Resource{
				resources.NewResource("TestName", "testZone1", earlyTime, reaperconfig.ResourceType_GCE_VM),
				resources.NewResource("TestingYetAnotherOne", "testZone1", earlyTime, reaperconfig.ResourceType_GCE_VM),
				resources.NewResource("TestThis", "testZone2", earlyTime, reaperconfig.ResourceType_GCE_VM),
			},
			"* * * * *",
		)...),
	},
	UpdateReaperTestCase{
		createReaperConfig(
			"project2", "", createResourceConfig(reaperconfig.ResourceType_GCE_VM, "Test", "Testing", "* * * * *", "testZone1", "testZone2"),
		),
		createTestReaper("project2", resources.CreateWatchlist(
			[]resources.Resource{
				resources.NewResource("TestName", "testZone1", earlyTime, reaperconfig.ResourceType_GCE_VM),
				resources.NewResource("TestThis", "testZone2", earlyTime, reaperconfig.ResourceType_GCE_VM),
			},
			"* * * * *",
		)...),
	},
	UpdateReaperTestCase{
		createReaperConfig(
			"project2", "", createResourceConfig(reaperconfig.ResourceType_GCE_VM, "Another", "", "* * * * *", "testZone1", "testZone2"),
		),
		createTestReaper("project2", resources.CreateWatchlist(
			[]resources.Resource{
				resources.NewResource("AnotherName", "testZone1", earlyTime, reaperconfig.ResourceType_GCE_VM),
				resources.NewResource("TestingYetAnotherOne", "testZone1", earlyTime, reaperconfig.ResourceType_GCE_VM),
				resources.NewResource("IsThisAnotherName", "testZone2", earlyTime, reaperconfig.ResourceType_GCE_VM),
			},
			"* * * * *",
		)...),
	},
}

func TestUpdateReaperConfig(t *testing.T) {
	server := createServer(getComputeEngineResourcesHandler)
	defer server.Close()

	testClientOptions := getTestClientOptions(server)
	testReaper := reaper.Reaper{}

	for _, testCase := range updateReaperTestCases {
		setupTestData()
		testReaper.UpdateReaperConfig(testContext, testCase.ReaperConfig, testClientOptions...)
		if !areWatchlistsEqual(testReaper, testCase.Expected) {
			t.Errorf("Reaper not updated correctly")
		}
	}
}

func createServer(handler http.HandlerFunc) *httptest.Server {
	// Use ServeMux if need to handle different endpoints
	// 	mux := http.NewServeMux()
	// 	mux.Handle("/", http.HandlerFunc(deleteResourceHandler))
	// 	// mux.Handle("/zones   /zone/instances", http.HandlerFunc(getResourcesHandler))
	// 	// mux.Handle("/", http.HandlerFunc(getComputeEngineResourcesHandler))
	// 	server := httptest.NewServer(mux)
	// 	return server
	return httptest.NewServer(http.HandlerFunc(handler))
}

func deleteComputeEngineResourceHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL.Path)
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
func areWatchlistsEqual(result, expected reaper.Reaper) bool {
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
		// Project to test which resources are deleted
		"project1": {
			reaperconfig.ResourceType_GCE_VM: {
				"testZone1": []TestData{
					TestData{"TestEarly", earlyTime.Format(time.RFC3339)},
					TestData{"TestFuture", lateTime.Format(time.RFC3339)},
				},
				"testZone2": []TestData{
					TestData{"TestMinuteAgo", minuteAgo.Format(time.RFC3339)},
					TestData{"TestMinuteLate", minuteLater.Format(time.RFC3339)},
				},
			},
		},
		// Project for testing updating watched resources
		"project2": {
			reaperconfig.ResourceType_GCE_VM: {
				"testZone1": []TestData{
					TestData{"TestName", earlyTime.Format(time.RFC3339)},
					TestData{"AnotherName", earlyTime.Format(time.RFC3339)},
					TestData{"TestingYetAnotherOne", earlyTime.Format(time.RFC3339)},
				},
				"testZone2": []TestData{
					TestData{"TestThis", earlyTime.Format(time.RFC3339)},
					TestData{"IsThisAnotherName", earlyTime.Format(time.RFC3339)},
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

func createTestReaper(projectID string, watchlist ...resources.WatchedResource) reaper.Reaper {
	return reaper.Reaper{
		UUID:      "TestUUID",
		ProjectID: projectID,
		Watchlist: watchlist,
		Schedule:  "TestSchedule",
	}
}
