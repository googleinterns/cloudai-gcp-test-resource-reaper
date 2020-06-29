package gcs

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/reaper"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/resources"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/utils"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
)

var (
	testInstances   map[string][]utils.TestInstance
	testTime        = "2020-06-17 10:00:00 -0400"
	deletedResource *utils.TestInstance
)

func TestAuth(t *testing.T) {
	client := &GCSClient{}
	err := client.Auth(context.Background())
	if err != nil {
		t.Errorf("GCS Auth failed with following error: %s", err.Error())
	}
}

func TestGetResources(t *testing.T) {
	server := utils.CreateServer(getResourcesHandler)
	defer server.Close()

	client := GCSClient{}
	client.Auth(context.TODO(), utils.GetTestOptions(server)...)
	config := reaper.NewResourceConfig(reaperconfig.ResourceType_GCE_VM, []string{"us", "us-east1"}, "test", "supercclank", "@every 1m")
	parsed, _ := client.GetResources("SampleProject1", config)
	fmt.Println(parsed)
}

type DeleteResourceTestCase struct {
	ProjectID string
	Name      string
	Expected  *utils.TestInstance
}

var deleteResourceTestCases = []DeleteResourceTestCase{
	DeleteResourceTestCase{"SampleProject1", "test-instance-1", &utils.TestInstance{"test-instance-1", testTime, "US"}},
	DeleteResourceTestCase{"SampleProject1", "test-instance-skip", &utils.TestInstance{"test-instance-skip", testTime, "NAM4"}},
	DeleteResourceTestCase{"SampleProject1", "wrong-name", nil},
	DeleteResourceTestCase{"SampleProject2", "another-test", &utils.TestInstance{"another-test", testTime, "NAM4"}},
	DeleteResourceTestCase{"SampleProject2", "another-wrong-name", nil},
	DeleteResourceTestCase{"SampleProject2", "", nil},
}

func TestDeleteResource(t *testing.T) {
	server := utils.CreateServer(deleteResourceHandler)
	defer server.Close()

	client := GCSClient{}
	client.Auth(context.TODO(), utils.GetTestOptions(server)...)

	for _, testCase := range deleteResourceTestCases {
		setupTestData()
		deletedResource = nil
		err := client.DeleteResource(testCase.ProjectID, resources.NewResource(testCase.Name, "TestZone", time.Now(), reaperconfig.ResourceType_GCS_BUCKET))
		if err != nil {
			t.Errorf("GCE Delete resource failed with the following error: %s", err.Error())
		}
		if !reflect.DeepEqual(testCase.Expected, deletedResource) {
			t.Error("Incorrect resource deleted")
		}
	}
}

func getResourcesHandler(w http.ResponseWriter, req *http.Request) {
	projectID := req.URL.Query()["project"][0]
	utils.SendResponse(w, testInstances[projectID])
	// w.Write([]byte(`{"success": true}`))
}

func deleteResourceHandler(w http.ResponseWriter, req *http.Request) {
	bucketName := strings.Split(req.URL.Path, "/")[2]
	for _, instances := range testInstances {
		for _, instance := range instances {
			// fmt.Println(instance.Name, bucketName)
			if strings.Compare(instance.Name, bucketName) == 0 {
				deletedResource = &utils.TestInstance{instance.Name, instance.CreationTimestamp, instance.Zone}
			}
		}
	}
	w.Write([]byte(`{"success": true}`))
}

func setupTestData() {
	testInstances = map[string][]utils.TestInstance{
		"SampleProject1": []utils.TestInstance{
			utils.TestInstance{"test-instance-1", testTime, "US"},
			utils.TestInstance{"test-instance-2", testTime, "US-EAST1"},
			utils.TestInstance{"test-instance-skip", testTime, "NAM4"},
			utils.TestInstance{"test-instance-another", testTime, "US-EAST1"},
		},
		"SampleProject2": []utils.TestInstance{
			utils.TestInstance{"another-instance", testTime, "US"},
			utils.TestInstance{"another-instance-skip", testTime, "US-EAST-1"},
			utils.TestInstance{"another-test", testTime, "NAM4"},
			utils.TestInstance{"another", testTime, "NAM4"},
		},
	}
}
