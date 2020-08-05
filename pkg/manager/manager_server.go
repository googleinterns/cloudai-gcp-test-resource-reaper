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

package manager

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/logger"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

// reaperManagerServer is the gRPC server for interacting with the reaper
// manager.
type reaperManagerServer struct {
	Manager       *ReaperManager
	clientOptions []option.ClientOption
}

// StartServer starts the gRPC server listing on the given address and port.
func StartServer(port string, clientOptions ...option.ClientOption) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}

	logger.Logf("------------------ Starting gRPC Server on :%s ------------------\n", port)
	defer logger.Log("------------------ Shutting down gRPC Server ------------------")

	server := grpc.NewServer()
	reaperconfig.RegisterReaperManagerServer(server, &reaperManagerServer{})
	server.Serve(lis)
}

// AddReaper adds a new reaper to the manager with the given config, and returns the UUID if the
// add was successful.
func (s *reaperManagerServer) AddReaper(ctx context.Context, config *reaperconfig.ReaperConfig) (*reaperconfig.Reaper, error) {
	if s.Manager == nil {
		return nil, fmt.Errorf("Reaper manager not started")
	}

	if watchedReaper := s.Manager.GetReaper(config.GetUuid()); watchedReaper != nil {
		err := fmt.Errorf("Reaper with UUID %s already exists", watchedReaper.UUID)
		return nil, err
	}
	s.Manager.AddReaperFromConfig(config)
	return &reaperconfig.Reaper{Uuid: config.GetUuid()}, nil
}

// UpdateReaper updates the reaper with the UUID given in the config with the data in the config, and returns
// the UUID if the update was successful.
func (s *reaperManagerServer) UpdateReaper(ctx context.Context, config *reaperconfig.ReaperConfig) (*reaperconfig.Reaper, error) {
	if s.Manager == nil {
		return nil, fmt.Errorf("Reaper manager not started")
	}

	if watchedReaper := s.Manager.GetReaper(config.GetUuid()); watchedReaper == nil {
		err := fmt.Errorf("Reaper with UUID %s does not exist", watchedReaper.UUID)
		return nil, err
	}
	s.Manager.UpdateReaper(config)
	return &reaperconfig.Reaper{Uuid: config.GetUuid()}, nil
}

// DeleteReaper deletes the reaper with the given UUID, and returns the UUID if the delete was successful.
func (s *reaperManagerServer) DeleteReaper(ctx context.Context, reaperToDelete *reaperconfig.Reaper) (*reaperconfig.Reaper, error) {
	if s.Manager == nil {
		return nil, fmt.Errorf("Reaper manager not started")
	}

	if watchedReaper := s.Manager.GetReaper(reaperToDelete.GetUuid()); watchedReaper == nil {
		err := fmt.Errorf("Reaper with UUID %s does not exist", reaperToDelete.GetUuid())
		return nil, err
	}
	s.Manager.DeleteReaper(reaperToDelete.GetUuid())
	return &reaperconfig.Reaper{Uuid: reaperToDelete.GetUuid()}, nil
}

// ListRunningReapers returns a list of UUIDs of all the running reapers.
func (s *reaperManagerServer) ListRunningReapers(ctx context.Context, req *empty.Empty) (*reaperconfig.ReaperCluster, error) {
	if s.Manager == nil {
		return nil, fmt.Errorf("Reaper manager not started")
	}

	reaperCluster := &reaperconfig.ReaperCluster{}
	for _, watchedReaper := range s.Manager.Reapers {
		reaper := &reaperconfig.Reaper{Uuid: watchedReaper.UUID}
		reaperCluster.Reapers = append(reaperCluster.Reapers, reaper)
	}
	return reaperCluster, nil
}

// StartManager begins the reaper manager process. This must be called before any reaper operations
// are invokved.
func (s *reaperManagerServer) StartManager(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	if s.Manager != nil {
		return new(empty.Empty), fmt.Errorf("reaper manager already running")
	}
	s.Manager = NewReaperManager(context.Background(), s.clientOptions...)
	go s.Manager.MonitorReapers()
	return new(empty.Empty), nil
}

// ShutdownManager ends the reaper manager process. This deletes all currently running reapers.
func (s *reaperManagerServer) ShutdownManager(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	if s.Manager == nil {
		return new(empty.Empty), fmt.Errorf("reaper manager already shutdown")
	}
	s.Manager.Shutdown()
	s.Manager = nil
	return new(empty.Empty), nil
}
