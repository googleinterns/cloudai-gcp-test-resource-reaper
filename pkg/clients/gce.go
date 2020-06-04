package clients

import (
	"context"

	gce "google.golang.org/api/compute/v1"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/resources"
)

func (client *GCEClient) Auth() error {
	ctx := context.Background()
	authedClient, err := gce.NewService(ctx)
	if err != nil {
		return err
	}
	client.Client = authedClient
	return nil
}

func (client *GCEClient) GetResource() resources.Resource {
	return resources.Resource{}
}

func (client *GCEClient) DeleteResource(resource resources.Resource) error {
	return nil
}
