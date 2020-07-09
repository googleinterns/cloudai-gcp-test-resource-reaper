package client

import (
	"context"
	"fmt"
	"log"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
	"google.golang.org/grpc"
)

type ReaperClient struct {
	client reaperconfig.ReaperManagerClient
	conn   *grpc.ClientConn
	ctx    context.Context
}

func StartClient(ctx context.Context, address, port string) *ReaperClient {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", address, port), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	reaperManagerClient := reaperconfig.NewReaperManagerClient(conn)

	client := &ReaperClient{}
	client.client = reaperManagerClient
	client.conn = conn
	client.ctx = ctx
	return client
}

func (c *ReaperClient) AddReaper(config *reaperconfig.ReaperConfig) {
	res, err := c.client.AddReaper(c.ctx, config)
	fmt.Println(res, err)
}

func (c *ReaperClient) UpdateReaper(config *reaperconfig.ReaperConfig) {
	res, err := c.client.UpdateReaper(c.ctx, config)
	fmt.Println(res, err)
}

func (c *ReaperClient) DeleteReaper(uuid string) {
	res, err := c.client.DeleteReaper(c.ctx, &reaperconfig.Reaper{Uuid: uuid})
	fmt.Println(res, err)
}

func (c *ReaperClient) ListRunningReapers() {
	res, err := c.client.ListRunningReapers(c.ctx, new(empty.Empty))
	fmt.Println(res, err)
}

func (c *ReaperClient) StartManager() {
	res, err := c.client.StartManager(c.ctx, new(empty.Empty))
	fmt.Println(res, err)
}

func (c *ReaperClient) ShutdownManager() {
	res, err := c.client.ShutdownManager(c.ctx, new(empty.Empty))
	fmt.Println(res, err)
}

func (c *ReaperClient) Close() {
	c.conn.Close()
}
