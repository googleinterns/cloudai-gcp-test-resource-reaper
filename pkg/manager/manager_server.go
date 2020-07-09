package manager

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
	"google.golang.org/grpc"
)

type reaperManagerServer struct {
	Manager *ReaperManager
}

func StartServer(address, port string) {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", address, port))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting gRPC Server on %s:%s\n", address, port)
	server := grpc.NewServer()
	reaperconfig.RegisterReaperManagerServer(server, &reaperManagerServer{})
	server.Serve(lis)
}

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

func (s *reaperManagerServer) StartManager(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	if s.Manager != nil {
		return new(empty.Empty), fmt.Errorf("reaper manager already running")
	}
	s.Manager = NewReaperManager(ctx)
	go s.Manager.MonitorReapers()
	return new(empty.Empty), nil
}

func (s *reaperManagerServer) ShutdownManager(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	if s.Manager == nil {
		return new(empty.Empty), fmt.Errorf("reaper manager already shutdown")
	}
	s.Manager.Shutdown()
	s.Manager = nil
	return new(empty.Empty), nil
}
