package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/resources"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
	"google.golang.org/api/option"
)

type Instance struct {
	Name              string
	Zone              string
	CreationTimestamp string
}

var testInstances = map[string]map[string][]Instance{
	"project1": {
		"testZone1": []Instance{
			newInstance("test1", "testZone1", "2019-10-12T07:20:50.52Z"),
			newInstance("test2", "testZone1", "2019-10-12T07:20:50.52Z"),
			newInstance("test3", "testZone1", "2019-10-12T07:20:50.52Z"),
			newInstance("differentName", "testZone1", "2019-10-12T07:20:50.52Z"),
		},
		"testZone2": []Instance{},
	},
}

func TestAuth(t *testing.T) {
	client := GCEClient{}
	client.Auth()

	computeAPIBaseURL := "https://compute.googleapis.com/compute/v1/projects/"
	if basePath := client.Client.BasePath; basePath != computeAPIBaseURL {
		t.Errorf("Base path = %s; want %s", basePath, computeAPIBaseURL)
	}
}

type GetResourceTestCase struct {
	ProjectID  string
	NameFilter string
	SkipFilter string
	Zones      []string
	Expected   []resources.Resource
}

var testGetResourcesCases = []GetResourceTestCase{
	GetResourceTestCase{"project1", "test", "", []string{"testZone1"}, []resources.Resource{}},
}

func TestGetResources(t *testing.T) {
	server := createServer(getResourcesHandler)
	defer server.Close()
	testClient := createTestGCEClient(server)

	for _, testCase := range testGetResourcesCases {
		config := reaperconfig.ResourceConfig{
			Zones:      testCase.Zones,
			NameFilter: testCase.NameFilter,
			SkipFilter: testCase.SkipFilter,
		}
		result, err := testClient.GetResources(testCase.ProjectID, config)
		if err != nil {
			t.Error(err)
		}
		fmt.Println(result)
		if !compareResourceLists(result, testCase.Expected) {
			// Improve this error message
			// Can't just print resource arrays since it is impossible to read
			t.Errorf("Resources not same as expected")
		}
	}
}

func TestDeleteResource(t *testing.T) {
	t.Fail()
}

type GetResourcesResponse struct {
	Items []Instance
}

func getResourcesHandler(w http.ResponseWriter, req *http.Request) {
	// Endpoint of the form: /{ProjectID}/zones/{ZoneName}/instances
	endpoint := req.URL.Path
	splitEndpoint := strings.Split(endpoint, "/")
	projectID := splitEndpoint[1]
	zone := splitEndpoint[3]

	res := GetResourcesResponse{testInstances[projectID][zone]}
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(res)
}

// O(n^2) runtime to compare slices and O(n) additional space, optimize?
func compareResourceLists(result, expected []resources.Resource) bool {
	usedResources := make([]bool, len(expected))
	var found bool
	for _, resultResource := range result {
		found = false
		for i, expectedResource := range expected {
			if usedResources[i] {
				continue
			}
			if reflect.DeepEqual(resultResource, expectedResource) {
				found = true
				usedResources[i] = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func deleteResourcesHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(*r)
}

func createServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handler))
}

func createTestGCEClient(server *httptest.Server) GCEClient {
	testOptions := []option.ClientOption{
		option.WithHTTPClient(server.Client()),
		option.WithEndpoint(server.URL),
	}

	gceClient := GCEClient{}
	gceClient.Auth(testOptions...)
	return gceClient
}

func newInstance(name, zone, creationTimestamp string) Instance {
	return Instance{
		name, zone, creationTimestamp,
	}
}
