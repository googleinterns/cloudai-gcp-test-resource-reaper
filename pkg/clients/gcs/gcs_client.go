// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gcs

import (
	"context"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/resources"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
	"google.golang.org/api/option"
)

// gcsBaseClient is common between GCS Buckets and GCS Objects.
type gcsBaseClient struct {
	client *storage.Client
	ctx    context.Context
}

// Auth authenticates the GCS client for both GCS Buckets and Objects.
func (client *gcsBaseClient) Auth(ctx context.Context, opts ...option.ClientOption) error {
	authedClient, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return err
	}
	client.client = authedClient
	client.ctx = ctx
	return nil
}

// GCSBucketClient is a client for GCS Buckets.
type GCSBucketClient struct {
	*gcsBaseClient
}

// NewGCSBucketClient creates a new GCS Bucket client.
func NewGCSBucketClient() *GCSBucketClient {
	return &GCSBucketClient{&gcsBaseClient{}}
}

// GetResources gets the GCS Bucket resources that match the given ResourceConfig.
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

// DeleteResource deletes the given GCS Bucket.
func (client *GCSBucketClient) DeleteResource(projectID string, resource *resources.Resource) error {
	bucketHandle := client.client.Bucket(resource.Name)
	err := bucketHandle.Delete(client.ctx)
	return err
}

// GCSObjectClient is a client for GCS objects. Note that the Zone
// for a GCS Object is the GCS Bucket name.
type GCSObjectClient struct {
	*gcsBaseClient
}

// NewGCSObjectClient creates a new GCS Object client.
func NewGCSObjectClient() *GCSObjectClient {
	return &GCSObjectClient{&gcsBaseClient{}}
}

// GetResources gets the GCS Object resources that match the given ResourceConfig.
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

// DeleteResource deletes the given GCS Object.
func (client *GCSObjectClient) DeleteResource(projectID string, resource *resources.Resource) error {
	bucketHandle := client.client.Bucket(resource.Zone)
	objectHandle := bucketHandle.Object(resource.Name)
	err := objectHandle.Delete(client.ctx)
	return err
}
