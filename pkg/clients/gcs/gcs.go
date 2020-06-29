package gcs

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/resources"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
	"google.golang.org/api/option"
)

type GCSClient struct {
	Client *storage.Client
	ctx    context.Context
}

func (client *GCSClient) Auth(ctx context.Context, opts ...option.ClientOption) error {
	authedClient, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return err
	}
	client.Client = authedClient
	client.ctx = ctx
	return nil
}

func (client *GCSClient) GetResources(projectID string, config *reaperconfig.ResourceConfig) ([]*resources.Resource, error) {
	var instances []*resources.Resource
	bucketIterator := client.Client.Buckets(client.ctx, projectID)
	for bucket, done := bucketIterator.Next(); done == nil; bucket, done = bucketIterator.Next() {
		bucketZone := bucket.Location
		for _, zone := range config.GetZones() {
			fmt.Println(zone)
			if strings.Compare(bucketZone, strings.ToUpper(zone)) != 0 {
				continue
			}
			name := bucket.Name
			fmt.Println(name)
			timeCreated := bucket.Created
			parsedResource := resources.NewResource(name, bucketZone, timeCreated, reaperconfig.ResourceType_GCS_BUCKET)
			if resources.ShouldAddResourceToWatchlist(parsedResource, config.GetNameFilter(), config.GetSkipFilter()) {
				instances = append(instances, parsedResource)
			}
		}
	}
	return instances, nil
}

func (client *GCSClient) DeleteResource(projectID string, resource *resources.Resource) error {
	bucketHandle := client.Client.Bucket(resource.Name)
	err := bucketHandle.Delete(client.ctx)
	return err
}
