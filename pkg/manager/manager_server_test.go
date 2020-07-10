package manager

import (
	"context"
	"log"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/reaper"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/utils"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var (
	lis         *bufconn.Listener
	testContext = context.Background()
)

func init() {
	mockServer := utils.CreateServer(utils.DefaultHandler)
	defer mockServer.Close()

	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	reaperconfig.RegisterReaperManagerServer(s, &reaperManagerServer{})
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

type AddReaperTestCase struct {
	Config   *reaperconfig.ReaperConfig
	Expected *reaperconfig.Reaper
}

var addReaperTestCases = []AddReaperTestCase{
	AddReaperTestCase{
		reaper.NewReaperConfig(nil, "@every 1h", "project", "TestReaper1"),
		&reaperconfig.Reaper{Uuid: "TestReaper1"},
	},
	AddReaperTestCase{
		reaper.NewReaperConfig(nil, "@every 1h", "project", "TestReaper2"),
		&reaperconfig.Reaper{Uuid: "TestReaper2"},
	},
	AddReaperTestCase{
		reaper.NewReaperConfig(nil, "@every 1h", "project", "TestReaper1"),
		nil,
	},
	AddReaperTestCase{
		reaper.NewReaperConfig(nil, "@every 1h", "project", "TestReaper3"),
		&reaperconfig.Reaper{Uuid: "TestReaper3"},
	},
}

func TestAddReaper(t *testing.T) {
	conn, err := grpc.DialContext(testContext, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := reaperconfig.NewReaperManagerClient(conn)

	client.StartManager(testContext, new(empty.Empty))
	defer client.ShutdownManager(testContext, new(empty.Empty))

	for _, testCase := range addReaperTestCases {
		resp, err := client.AddReaper(testContext, testCase.Config)
		if testCase.Expected != nil {
			if err != nil {
				t.Fatalf("Add reaper failed: %v", err)
			} else if strings.Compare(testCase.Expected.Uuid, testCase.Config.Uuid) != 0 {
				t.Fatalf("Got UUID %s, expected %s", testCase.Expected.Uuid, resp.Uuid)
			}
		} else if testCase.Expected == nil && err == nil {
			t.Fatalf("Expected error to be thrown since name already exists")
		}
		time.Sleep(time.Second * 2)
	}
}

type UpdateReaperTestCase struct {
	Config   *reaperconfig.ReaperConfig
	Expected *reaperconfig.Reaper
}

var updateReaperTestCases = []UpdateReaperTestCase{
	UpdateReaperTestCase{},
}

func TestUpdateReaper(t *testing.T) {
	conn, err := grpc.DialContext(testContext, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := reaperconfig.NewReaperManagerClient(conn)

	for _, testCase := range updateReaperTestCases {
		resp, err := client.UpdateReaper(testContext, testCase.Config)
		if err != nil {
			t.Fatalf("Error")
		}
		if resp != testCase.Expected {
			t.Fatalf("Wrong")
		}
	}

}

type DeleteReaperTestCase struct {
	UUID     string
	Expected *reaperconfig.Reaper
}

var deleteReaperTestCases = []DeleteReaperTestCase{
	DeleteReaperTestCase{},
}

func TestDeleteReaper(t *testing.T) {
	conn, err := grpc.DialContext(testContext, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := reaperconfig.NewReaperManagerClient(conn)

	for _, testCase := range deleteReaperTestCases {
		testReaper := &reaperconfig.Reaper{Uuid: testCase.UUID}
		resp, err := client.DeleteReaper(testContext, testReaper)
		if err != nil {
			t.Fatalf("Error")
		}
		if resp != testCase.Expected {
			t.Fatalf("Wrong")
		}
	}
}

var listReapersTestCases = [][]string{}

// func TestListRunningReapers(t *testing.T) {
// 	conn, err := grpc.DialContext(testContext, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
// 	if err != nil {
// 		t.Fatalf("Failed to dial bufnet: %v", err)
// 	}
// 	defer conn.Close()
// 	client := reaperconfig.NewReaperManagerClient(conn)

// 	for _, uuids := range listReapersTestCases {

// 	}
// }

func TestStartManager(t *testing.T) {
	conn, err := grpc.DialContext(testContext, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := reaperconfig.NewReaperManagerClient(conn)

	_, err = client.StartManager(testContext, new(empty.Empty))
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	_, err = client.StartManager(testContext, new(empty.Empty))
	if err == nil {
		t.Fatalf("Manager should already by running, and throw an error")
	}
}

func TestShutdownManager(t *testing.T) {
	conn, err := grpc.DialContext(testContext, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := reaperconfig.NewReaperManagerClient(conn)

	_, err = client.ShutdownManager(testContext, new(empty.Empty))
	if err == nil {
		t.Fatalf("Manager should not be running, and throw an error when shutting down")
	}

	client.StartManager(testContext, new(empty.Empty))
	_, err = client.ShutdownManager(testContext, new(empty.Empty))
	if err != nil {
		t.Fatalf("Failed to shutdown manager: %v", err)
	}

	_, err = client.ShutdownManager(testContext, new(empty.Empty))
	if err == nil {
		t.Fatalf("Manager should not be running, and throw an error when shutting down")
	}
}
