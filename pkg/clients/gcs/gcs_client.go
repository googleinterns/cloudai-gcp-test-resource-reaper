package gcs

import (
	"context"
	"fmt"
	"strings"

	"errors"

	"cloud.google.com/go/storage"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/resources"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
	"google.golang.org/api/option"
)

type gcsBaseClient struct {
	Client *storage.Client
	ctx    context.Context
}

func (client *gcsBaseClient) Auth(ctx context.Context, opts ...option.ClientOption) error {
	authedClient, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return err
	}
	client.Client = authedClient
	client.ctx = ctx
	return nil
}

type GCSBucketClient struct {
	*gcsBaseClient
}

func (client *GCSBucketClient) GetResources(projectID string, config *reaperconfig.ResourceConfig) ([]*resources.Resource, error) {
	var instances []*resources.Resource
	bucketIterator := client.Client.Buckets(client.ctx, projectID)
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
	bucketHandle := client.Client.Bucket(resource.Name)
	err := bucketHandle.Delete(client.ctx)
	return err
}

// Name filter for a GCSObject should be of the form [BucketName]:[ObjectNameRegex]
type GCSObjectClient struct {
	*gcsBaseClient
}

func (client *GCSObjectClient) GetResources(projectID string, config *reaperconfig.ResourceConfig) ([]*resources.Resource, error) {
	var instances []*resources.Resource
	bucketName, objectRegex, err := parseObjectName(config.GetNameFilter())
	if err != nil {
		return nil, err
	}
	bucketHandle := client.Client.Bucket(bucketName)

	// https://pkg.go.dev/cloud.google.com/go/storage?tab=doc#Query
	var query *storage.Query = nil
	objectIterator := bucketHandle.Objects(client.ctx, query)
	for object, done := objectIterator.Next(); done == nil; object, done = objectIterator.Next() {
		fmt.Println("OBJECT: ", object.Name)
		// fmt.Println(resources.ShouldAddResourceToWatchlist())
	}
	fmt.Println(objectRegex)
	return instances, nil
}

func (client *GCSObjectClient) DeleteResource(projectID string, resource *resources.Resource) error {
	return nil
}

// Assuming name is of the format BucketName:ObjectNameRegex
func parseObjectName(name string) (string, string, error) {
	splitName := strings.Split(name, ":")
	if len(splitName) != 2 {
		return "", "", errors.New("gcs object name filter should be formatted in the form [bucketName]:[objectNameRegex]")
	}
	return splitName[0], splitName[1], nil
}
