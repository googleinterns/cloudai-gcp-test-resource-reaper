package gcs

import (
	"context"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/resources"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
	"google.golang.org/api/option"
)

type gcsBaseClient struct {
	client *storage.Client
	ctx    context.Context
}

func (client *gcsBaseClient) Auth(ctx context.Context, opts ...option.ClientOption) error {
	authedClient, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return err
	}
	client.client = authedClient
	client.ctx = ctx
	return nil
}

type GCSBucketClient struct {
	*gcsBaseClient
}

func NewGCSBucketClient() *GCSBucketClient {
	return &GCSBucketClient{&gcsBaseClient{}}
}

func (client *GCSBucketClient) GetResources(projectID string, config *reaperconfig.ResourceConfig) ([]*resources.Resource, error) {
	var instances []*resources.Resource
	bucketIterator := client.client.Buckets(client.ctx, projectID)
	for bucket, done := bucketIterator.Next(); done == nil; bucket, done = bucketIterator.Next() {
		bucketZone := bucket.Location
		for _, zone := range config.GetZones() {
			if strings.Compare(bucketZone, strings.ToUpper(zone)) != 0 {
				continue
			}
			name := bucket.Name
			timeCreated := bucket.Created
			parsedResource := resources.NewResource(name, bucketZone, timeCreated, reaperconfig.ResourceType_GCS_BUCKET)
			if resources.ShouldAddResourceToWatchlist(parsedResource, config.GetNameFilter(), config.GetSkipFilter()) {
				instances = append(instances, parsedResource)
			}
		}
	}
	return instances, nil
}

func (client *GCSBucketClient) DeleteResource(projectID string, resource *resources.Resource) error {
	bucketHandle := client.client.Bucket(resource.Name)
	err := bucketHandle.Delete(client.ctx)
	return err
}

// Zone for a GCSObject is the name of the bucket (bucket names have to be globally unique)
type GCSObjectClient struct {
	*gcsBaseClient
}

func NewGCSObjectClient() *GCSObjectClient {
	return &GCSObjectClient{&gcsBaseClient{}}
}

func (client *GCSObjectClient) GetResources(projectID string, config *reaperconfig.ResourceConfig) ([]*resources.Resource, error) {
	var instances []*resources.Resource

	for _, bucket := range config.GetZones() {
		bucketHandle := client.client.Bucket(bucket)
		objectIterator := bucketHandle.Objects(client.ctx, nil)

		for object, done := objectIterator.Next(); done == nil; object, done = objectIterator.Next() {
			objectResource := resources.NewResource(object.Name, bucket, object.Created, reaperconfig.ResourceType_GCS_OBJECT)
			if resources.ShouldAddResourceToWatchlist(objectResource, config.GetNameFilter(), config.GetSkipFilter()) {
				instances = append(instances, objectResource)
			}
		}
	}
	return instances, nil
}

func (client *GCSObjectClient) DeleteResource(projectID string, resource *resources.Resource) error {
	bucketHandle := client.client.Bucket(resource.Zone)
	objectHandle := bucketHandle.Object(resource.Name)
	err := objectHandle.Delete(client.ctx)
	return err
}
