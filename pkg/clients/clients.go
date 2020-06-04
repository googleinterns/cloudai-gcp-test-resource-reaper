package clients

import (
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/resources"
	gce "google.golang.org/api/compute/v1"
)

type Client interface {
	Auth() error
	GetResource() resources.Resource
	DeleteResource(resources.Resource) error
}

type GCEClient struct {
	Client *gce.Service
}
