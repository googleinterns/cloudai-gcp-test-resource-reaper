package clients

import (
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/resources"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
	gce "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

type Client interface {
	Auth(opts ...option.ClientOption) error
	GetResources(projectID string, config reaperconfig.ResourceConfig) ([]resources.Resource, error)
	DeleteResource(projectID string, resource resources.Resource) error
}

type GCEClient struct {
	Client *gce.Service
}
