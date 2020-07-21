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

package client

import (
	"context"
	"fmt"
	"log"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
	"google.golang.org/grpc"
)

// ReaperClient is a gRPC client for communicating with the reaper manager server.
type ReaperClient struct {
	client reaperconfig.ReaperManagerClient
	conn   *grpc.ClientConn
	ctx    context.Context
}

// StartClient returns a client with a gRPC connection with the server running on address:port.
func StartClient(ctx context.Context, address, port string) *ReaperClient {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", address, port), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	reaperManagerClient := reaperconfig.NewReaperManagerClient(conn)

	return &ReaperClient{
		client: reaperManagerClient,
		conn:   conn,
		ctx:    ctx,
	}
}

// AddReaper adds a new reaper the reaper manager.
func (c *ReaperClient) AddReaper(config *reaperconfig.ReaperConfig) (string, error) {
	res, err := c.client.AddReaper(c.ctx, config)
	return res.Uuid, err
}

// UpdateReaper updates the reaper from the given ReaperConfig. If the UUID in the config does not exist,
// an error will be thrown.
func (c *ReaperClient) UpdateReaper(config *reaperconfig.ReaperConfig) (string, error) {
	res, err := c.client.UpdateReaper(c.ctx, config)
	return res.Uuid, err
}

// DeleteReaper deletes the reaper with the given UUID.
func (c *ReaperClient) DeleteReaper(uuid string) error {
	_, err := c.client.DeleteReaper(c.ctx, &reaperconfig.Reaper{Uuid: uuid})
	return err
}

// ListRunningReapers returns a list of the running reapers' UUIDs.
func (c *ReaperClient) ListRunningReapers() ([]string, error) {
	res, err := c.client.ListRunningReapers(c.ctx, new(empty.Empty))
	if err != nil {
		return nil, err
	}
	var runningReapers []string
	for _, reaper := range res.Reapers {
		runningReapers = append(runningReapers, reaper.Uuid)
	}
	return runningReapers, nil
}

// StartManager starts running the reaper manager. Note this is different from starting the gRPC
// server.
func (c *ReaperClient) StartManager() error {
	_, err := c.client.StartManager(c.ctx, new(empty.Empty))
	return err
}

// ShutdownManager stops the reaper manager process. Note this does not shut down the gRPC server.
func (c *ReaperClient) ShutdownManager() error {
	_, err := c.client.ShutdownManager(c.ctx, new(empty.Empty))
	return err
}

// Close closes the ReaperClient connection to the gRPC server.
func (c *ReaperClient) Close() {
	c.conn.Close()
}
