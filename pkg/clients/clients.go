package clients

import (
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/resources"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
	gce "google.golang.org/api/compute/v1"
)

type Client interface {
	Auth() error
	GetResources(projectID string, config reaperconfig.ResourceConfig) ([]resources.Resource, error)
	DeleteResource(projectID string, resource resources.Resource) error
}

type GCEClient struct {
	Client *gce.Service
}
