package reaper

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
)

var (
	projectID   string
	accessToken string
	ctx         = context.Background()
)

func TestReaperIntegration(t *testing.T) {
	setup(true)
	resources := []*reaperconfig.ResourceConfig{
		NewResourceConfig(reaperconfig.ResourceType_GCE_VM, []string{"us-east1-b", "us-east1-c"}, "test", "", "9 7 * * *"),
		NewResourceConfig(reaperconfig.ResourceType_GCE_VM, []string{"us-east1-b"}, "Another", "", "1 * * * *"),
	}
	reaperConfig := NewReaperConfig(resources, "TestSchedule", "SkipFilter", projectID, "UUID")

	reaper := NewReaper()
	reaper.UpdateReaperConfig(ctx, reaperConfig)

	// Set current time to 10 years later for testing
	reaper.freezeTime(time.Now().AddDate(10, 0, 0))

	reaper.PrintWatchlist()
	// reaper.RunThroughResources(ctx)
	// reaper.PrintWatchlist()
}

type TestConfig struct {
	ProjectID   string `json:"projectId"`
	AccessToken string `json:"accessToken"`
}

func setup(shouldCreateResources bool) {
	readConfigFile()
	if shouldCreateResources {
		createTestResources()
	}
}

func readConfigFile() {
	var configData TestConfig
	jsonConfigFile, err := os.Open("config.json")
	defer jsonConfigFile.Close()
	if err != nil {
		log.Fatal(err)
	}
	configParser := json.NewDecoder(jsonConfigFile)
	configParser.Decode(&configData)

	projectID = configData.ProjectID
	accessToken = configData.AccessToken
}

type TestResource struct {
	Name     string
	Zone     string
	DiskName string
}

var testResources = []TestResource{
	TestResource{"test-resource-1", "us-east1-b", "test-disk-2"},
	TestResource{"test-resource-2", "us-east1-b", "test-disk-2"},
	TestResource{"test-resource-3", "us-east1-c", "test-disk-2"},
	TestResource{"another-resource-1", "us-east1-b", "test-disk-4"},
	TestResource{"another-resource-2", "us-east1-b", "test-disk-5"},
}

func createTestResources() {
	for _, resource := range testResources {
		createGCEInstance(ctx, resource.Name, resource.Zone, resource.DiskName)
	}
}

func createGCEInstance(ctx context.Context, name, zone, diskName string) {
	endpoint := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s/instances", projectID, zone)

	reqBody := struct {
		MachineType       string `json:"machineType"`
		Name              string `json:"name"`
		NetworkInterfaces []struct {
			Network string `json:"network"`
		} `json:"networkInterfaces"`
		Disks []struct {
			Boot             bool `json:"boot"`
			AutoDelete       bool `json:"autoDelete"`
			InitializeParams struct {
				DiskName    string `json:"diskName"`
				SourceImage string `json:"sourceImage"`
			} `json:"initializeParams"`
			Mode      string `json:"mode"`
			Interface string `json:"interface"`
		} `json:"disks"`
	}{
		Name:        name,
		MachineType: fmt.Sprintf("zones/%s/machineTypes/f1-micro", zone),

		NetworkInterfaces: []struct {
			Network string `json:"network"`
		}{
			{
				Network: fmt.Sprintf("projects/%s/global/networks/default", projectID),
			},
		}, //For simplicity use the default network
		Disks: []struct {
			Boot             bool `json:"boot"`
			AutoDelete       bool `json:"autoDelete"`
			InitializeParams struct {
				DiskName    string `json:"diskName"`
				SourceImage string `json:"sourceImage"`
			} `json:"initializeParams"`
			Mode      string `json:"mode"`
			Interface string `json:"interface"`
		}{
			{
				Boot:       true,
				AutoDelete: false,
				Mode:       "READ_WRITE",
				Interface:  "SCSI",
				InitializeParams: struct {
					DiskName    string `json:"diskName"`
					SourceImage string `json:"sourceImage"`
				}{
					DiskName:    "test-disk-2",
					SourceImage: "projects/debian-cloud/global/images/family/debian-9",
				},
			},
		},
	}

	bodyData, err := json.Marshal(reqBody)
	if err != nil {
		log.Println(err.Error())
	}

	request, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(string(bodyData)))
	if err != nil {
		log.Println(err.Error())
	}
	request.Header.Set(http.CanonicalHeaderKey("authorization"), fmt.Sprintf("Bearer %s", accessToken))
	request.Header.Set(http.CanonicalHeaderKey("content-type"), "application/json")

	client := http.DefaultClient
	response, err := client.Do(request)
	if err != nil {
		log.Println(err.Error())
	}

	data, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		log.Println(err.Error())
	}

	fmt.Println(string(data))
}
