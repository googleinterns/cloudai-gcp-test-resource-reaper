package clients

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
	"google.golang.org/api/option"
)

func TestAuth(t *testing.T) {
	client := GCEClient{}
	client.Auth()

	computeAPIBaseURL := "https://compute.googleapis.com/compute/v1/projects/"
	if basePath := client.Client.BasePath; basePath != computeAPIBaseURL {
		t.Errorf("Base path = %s; want %s", basePath, computeAPIBaseURL)
	}
}

func TestGetResources(t *testing.T) {
	server := createServer(listResourcesHandler)
	defer server.Close()
	testClient := createTestGCEClient(server)

	
	testClient.GetResources("", reaperconfig.ResourceConfig{Zones: []string{"Zone1"}})
}

func TestDeleteResource(t *testing.T) {
	t.Fail()
}

func listResourcesHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(*r)
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
