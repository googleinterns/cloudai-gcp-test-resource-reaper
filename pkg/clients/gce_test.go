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

package clients

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/resources"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
	"google.golang.org/api/option"
)

// A mock object to represent a Compute Engine Instance. Only the name
// and creation time are needed for testing the client.
type Instance struct {
	Name              string
	CreationTimestamp string
}

var (
	// The time created is not important for client testing purposes,
	// however is a required field, so a random value was chosen and
	// used throughout the tests.
	timeCreatedString = "2019-10-12T07:20:50.52Z"
	timeCreated, _    = time.Parse(time.RFC3339, timeCreatedString)

	// Map of project -> Zones in project -> Instances in Zone. This
	// mocks the data that would be stored in GCP. Call setupTestInstances
	// to populate with data.
	testInstances map[string]map[string][]Instance
)

// TestAuth tests the authentication method of the Compute Engine Client.
func TestAuth(t *testing.T) {
	client := GCEClient{}
	client.Auth()

	computeAPIBaseURL := "https://compute.googleapis.com/compute/v1/projects/"
	if basePath := client.Client.BasePath; basePath != computeAPIBaseURL {
		t.Errorf("Base path = %s; want %s", basePath, computeAPIBaseURL)
	}
}

// A GetResourcesTestCase is a struct for organizing test inputs and expected outputs
// for testing client's the GetResources method.
type GetResourcesTestCase struct {
	ProjectID  string
	NameFilter string
	SkipFilter string
	Zones      []string
	Expected   []resources.Resource
}

// The test cases for GetResources method
var testGetResourcesCases = []GetResourcesTestCase{
	GetResourcesTestCase{"project1", "test", "", []string{"testZone1"}, []resources.Resource{
		resources.NewResource("test1", "testZone1", timeCreated, reaperconfig.ResourceType_GCE_VM),
		resources.NewResource("test2", "testZone1", timeCreated, reaperconfig.ResourceType_GCE_VM),
		resources.NewResource("test3", "testZone1", timeCreated, reaperconfig.ResourceType_GCE_VM),
	}},
	GetResourcesTestCase{"project1", "test", "", []string{"testZone1", "testZone2"}, []resources.Resource{
		resources.NewResource("test1", "testZone1", timeCreated, reaperconfig.ResourceType_GCE_VM),
		resources.NewResource("test2", "testZone1", timeCreated, reaperconfig.ResourceType_GCE_VM),
		resources.NewResource("test3", "testZone1", timeCreated, reaperconfig.ResourceType_GCE_VM),
		resources.NewResource("test1", "testZone2", timeCreated, reaperconfig.ResourceType_GCE_VM),
		resources.NewResource("test2", "testZone2", timeCreated, reaperconfig.ResourceType_GCE_VM),
	}},
	GetResourcesTestCase{"project1", "test", "test1", []string{"testZone1"}, []resources.Resource{
		resources.NewResource("test2", "testZone1", timeCreated, reaperconfig.ResourceType_GCE_VM),
		resources.NewResource("test3", "testZone1", timeCreated, reaperconfig.ResourceType_GCE_VM),
	}},
	GetResourcesTestCase{"project1", "test", "test1", []string{"testZone1", "testZone2"}, []resources.Resource{
		resources.NewResource("test2", "testZone1", timeCreated, reaperconfig.ResourceType_GCE_VM),
		resources.NewResource("test3", "testZone1", timeCreated, reaperconfig.ResourceType_GCE_VM),
		resources.NewResource("test2", "testZone2", timeCreated, reaperconfig.ResourceType_GCE_VM),
	}},
	GetResourcesTestCase{"project1", "different", "", []string{"testZone1"}, []resources.Resource{
		resources.NewResource("differentName", "testZone1", timeCreated, reaperconfig.ResourceType_GCE_VM),
	}},
	GetResourcesTestCase{"project2", "test", "", []string{"testZone1"}, []resources.Resource{
		resources.NewResource("testProject2", "testZone1", timeCreated, reaperconfig.ResourceType_GCE_VM),
	}},
}

// TestGetResources tests the clients GetResources method. Note that the order in which the resources
// are returned from the method does not matter, and this only tests whether the resources returned are
// equal in value and number to what is expected.
func TestGetResources(t *testing.T) {
	server := createServer(getResourcesHandler)
	defer server.Close()
	testClient := createTestGCEClient(server)

	setupManyTestInstances()
	for _, testCase := range testGetResourcesCases {
		config := &reaperconfig.ResourceConfig{
			Zones:      testCase.Zones,
			NameFilter: testCase.NameFilter,
			SkipFilter: testCase.SkipFilter,
		}
		result, err := testClient.GetResources(testCase.ProjectID, config)
		if err != nil {
			t.Error(err)
		}
		if !compareResourceLists(result, testCase.Expected) {
			// Improve this error message
			// Can't just print resource arrays since it is impossible to read
			t.Errorf("Resources not same as expected")
		}
	}
}

// A DeleteResourceTestCase is a struct for organizing test inputs and expected outputs
// for testing client's the DeleteResource method.
type DeleteResourceTestCase struct {
	ProjectID string
	Resource  resources.Resource
	Expected  map[string]map[string][]Instance
}

// Test cases for DeleteResource.
var testDeleteResourceCases = []DeleteResourceTestCase{
	DeleteResourceTestCase{
		"project1",
		resources.NewResource("test", "testZone1", timeCreated, reaperconfig.ResourceType_GCE_VM),
		map[string]map[string][]Instance{
			"project1": {
				"testZone1": []Instance{},
				"testZone2": []Instance{
					newInstance("test", timeCreatedString),
				},
			},
		},
	},
	DeleteResourceTestCase{
		"project1",
		resources.NewResource("test", "testZone2", timeCreated, reaperconfig.ResourceType_GCE_VM),
		map[string]map[string][]Instance{
			"project1": {
				"testZone1": []Instance{
					newInstance("test", timeCreatedString),
				},
				"testZone2": []Instance{},
			},
		},
	},
}

// TestDeleteResource tests the Compute Engine client's DeleteResource method.
func TestDeleteResource(t *testing.T) {
	server := createServer(deleteResourcesHandler)
	defer server.Close()
	testClient := createTestGCEClient(server)

	for _, testCase := range testDeleteResourceCases {
		setupFewTestInstances()
		err := testClient.DeleteResource(testCase.ProjectID, testCase.Resource)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(testInstances, testCase.Expected) {
			t.Errorf("Delete not working correctly")
		}
	}
}

type GetResourcesResponse struct {
	Items []Instance
}

// Mock server's http handler for GetResources test
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

type DeleteResourceResponse struct {
	Text string
}

// Mock server's http handler for DeleteResource test
func deleteResourcesHandler(w http.ResponseWriter, req *http.Request) {
	// Endpoint of the form: /{ProjectID}/zones/{ZoneName}/instances/{InstanceName}
	endpoint := req.URL.Path
	splitEndpoint := strings.Split(endpoint, "/")
	projectID := splitEndpoint[1]
	zone := splitEndpoint[3]
	name := splitEndpoint[5]

	var index int
	instancesInZone := testInstances[projectID][zone]
	for i, instance := range instancesInZone {
		if strings.Compare(instance.Name, name) == 0 {
			index = i
			break
		}
	}
	// Removing instance with name match in zone and project
	testInstances[projectID][zone] = append(instancesInZone[:index], instancesInZone[index+1:]...)

	res := DeleteResourceResponse{"Successfully Deleted"}
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(res)

}

// compareResourceLists compares the resources returned from a GetResources
// call to what was expected.
func compareResourceLists(result, expected []resources.Resource) bool {
	if len(result) != len(expected) {
		return false
	}
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

// createServer is a helper function to create a fake server
// where http requsts will be rerouted for testing
func createServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handler))
}

// createTestGCEClient creates a Compute Engine client that sends
// all http requests to the fake server.
func createTestGCEClient(server *httptest.Server) GCEClient {
	testOptions := []option.ClientOption{
		option.WithHTTPClient(server.Client()),
		option.WithEndpoint(server.URL),
	}

	gceTestClient := GCEClient{}
	gceTestClient.Auth(testOptions...)
	return gceTestClient
}

// newInstance constructs an Instance struct.
func newInstance(name, creationTimestamp string) Instance {
	return Instance{
		name, creationTimestamp,
	}
}

// Populate testInstances with lots of data for GetResources test
func setupManyTestInstances() {
	testInstances = map[string]map[string][]Instance{
		"project1": {
			"testZone1": []Instance{
				newInstance("test1", timeCreatedString),
				newInstance("test2", timeCreatedString),
				newInstance("test3", timeCreatedString),
				newInstance("differentName", timeCreatedString),
			},
			"testZone2": []Instance{
				newInstance("test1", timeCreatedString),
				newInstance("test2", timeCreatedString),
			},
		},
		"project2": {
			"testZone1": []Instance{
				newInstance("testProject2", timeCreatedString),
				newInstance("different", timeCreatedString),
			},
		},
	}
}

// Populate testInstances with small amount of data for DeleteResource test
func setupFewTestInstances() {
	testInstances = map[string]map[string][]Instance{
		"project1": {
			"testZone1": []Instance{
				newInstance("test", timeCreatedString),
			},
			"testZone2": []Instance{
				newInstance("test", timeCreatedString),
			},
		},
	}
}
