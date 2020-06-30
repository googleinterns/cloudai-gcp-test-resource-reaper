package gcs

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/resources"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/utils"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
)

// Maybe look here for testing:
// https://github.com/googleapis/google-cloud-go-testing
// https://github.com/googleapis/google-cloud-go/issues/592#issuecomment-406099221

var (
	testInstances   map[string][]utils.TestInstance
	testTime        = "2020-06-17 10:00:00 -0400"
	deletedResource *utils.TestInstance
)

func TestAuth(t *testing.T) {
	bucketClient := NewGCSBucketClient()
	err := bucketClient.Auth(context.TODO())
	if err != nil {
		t.Errorf("GCS Bucket Auth failed with following error: %s", err.Error())
	}
	objectClient := NewGCSObjectClient()
	err = objectClient.Auth(context.TODO())
	if err != nil {
		t.Errorf("GCS Object Auth failed with following error: %s", err.Error())
	}
}

// func TestGetBucketResources(t *testing.T) {
// 	server := utils.CreateServer(getResourcesHandler)
// 	defer server.Close()

// 	client := GCSBucketClient{}
// 	client.Auth(context.TODO(), utils.GetTestOptions(server)...)
// 	config := reaper.NewResourceConfig(reaperconfig.ResourceType_GCE_VM, []string{"us", "us-east1"}, "test", "supercclank", "@every 1m")
// 	parsed, _ := client.GetResources("SampleProject1", config)
// 	fmt.Println(parsed)
// }

type DeleteBucketResourceTestCase struct {
	ProjectID string
	Name      string
	Expected  *utils.TestInstance
}

var deleteBucketResourceTestCases = []DeleteBucketResourceTestCase{
	DeleteBucketResourceTestCase{"SampleProject1", "test-instance-1", &utils.TestInstance{"test-instance-1", testTime, "US"}},
	DeleteBucketResourceTestCase{"SampleProject1", "test-instance-skip", &utils.TestInstance{"test-instance-skip", testTime, "NAM4"}},
	DeleteBucketResourceTestCase{"SampleProject1", "wrong-name", nil},
	DeleteBucketResourceTestCase{"SampleProject2", "another-test", &utils.TestInstance{"another-test", testTime, "NAM4"}},
	DeleteBucketResourceTestCase{"SampleProject2", "another-wrong-name", nil},
	DeleteBucketResourceTestCase{"SampleProject2", "", nil},
}

func TestDeleteBucketResource(t *testing.T) {
	server := utils.CreateServer(deleteBucketResourceHandler)
	defer server.Close()

	client := NewGCSBucketClient()
	client.Auth(context.TODO(), utils.GetTestOptions(server)...)

	for _, testCase := range deleteBucketResourceTestCases {
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

func TestGetObjectResource(t *testing.T) {

}

func TestDeleteObjectResource(t *testing.T) {

}

func getResourcesHandler(w http.ResponseWriter, req *http.Request) {
	projectID := req.URL.Query()["project"][0]
	utils.SendResponse(w, testInstances[projectID])
	// w.Write([]byte(`{"success": true}`))
}

func deleteBucketResourceHandler(w http.ResponseWriter, req *http.Request) {
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
