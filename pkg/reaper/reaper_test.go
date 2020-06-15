package reaper_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/reaper"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
	"google.golang.org/api/option"
	"google.golang.org/api/internal"
)

var (
	testContext = context.Background()
	earlyTime   = "2010-10-12T07:20:50.52Z"
	// ProjectID -> ResourceType -> Zones -> Resource
	testData        map[string]map[reaperconfig.ResourceType]map[string][]TestData
	testTimeCreated = time.Now()
)

func TestRunThroughResources(t *testing.T) {
	fmt.Println(time.Now().String())
}

type UpdateReaperTestCase struct {
	ReaperConfig *reaperconfig.ReaperConfig
	Expected     reaper.Reaper
}

var updateReaperTestCases = []UpdateReaperTestCase{
	UpdateReaperTestCase{},
}

func TestUpdateReaperConfig(t *testing.T) {
	server := createServer()
	defer server.Close()

	// testClient := getTestClient(server)
	clientOptions := getTestClientOptions(server)
	testReaper := reaper.Reaper{}

	for _, testCase := range updateReaperTestCases {
		testReaper.UpdateReaperConfig(testContext, testCase.ReaperConfig, clientOptions...)
		if !reflect.DeepEqual(testReaper, testCase.Expected) {
			t.Errorf("Reaper not updated correctly")
		}
	}
}

func createServer() *httptest.Server {
	mux := http.NewServeMux()
	// mux.Handle("/delete", http.HandlerFunc(deleteResourceHandler))
	// mux.Handle("/getResources", http.HandlerFunc(getResourcesHandler))
	mux.Handle("/", http.HandlerFunc(getResourcesHandler))
	server := httptest.NewServer(mux)
	return server
}

func deleteResourceHandler(w http.ResponseWriter, req *http.Request) {

}

func getResourcesHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL)
}

// func newTestClient(server *httptest.Server) *TestClient {
// 	return &TestClient{server.Client()}
// }

func getTestClientOptions(server *httptest.Server) []option.ClientOption {
	return []option.ClientOption{
		option.WithHTTPClient(server.Client()),
		option.WithEndpoint(server.URL),
	}
}

// type TestClient struct {
// 	*http.Client
// }

// Implementing these makes http.client implement Client interface
// func (client *http.Client) Auth(ctx context.Context, opts ...option.ClientOption) error {
// 	fmt.Println("IMPLEMENTED AUTH")
// 	return nil
// }

// func (client *http.Client) GetResources(projectID string, config *reaperconfig.ResourceConfig) ([]resources.Resource, error) {
// 	fmt.Println("IMPLEMENTED GET")
// 	return nil, nil
// }

// func (client *http.Client) DeleteResource(projectID string, resource resources.Resource) error {
// 	fmt.Println("IMPLEMENTED DELETE")
// 	return nil
// }

type TestData struct {
	Name              string
	CreationTimestamp string
}

func setupTestData() {
	testData = map[string]map[reaperconfig.ResourceType]map[string][]TestData{
		"project1": {
			reaperconfig.ResourceType_GCE_VM: {
				"testZone1": []TestData{
					TestData{"TestName", earlyTime},
				},
				"testZone2": []TestData{
					TestData{},
				},
			},
		},
		"project2": {},
	}
}
