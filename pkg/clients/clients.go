package clients

import (
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/resources"
	reaperpb "github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
	gce "google.golang.org/api/compute/v1"
)

type Client interface {
	Auth() error
	GetResources(projectID string, config reaperpb.ResourceConfig) resources.Resource
	DeleteResource(resource resources.Resource) error
}

type GCEClient struct {
	Client *gce.Service
}
