package client

import (
	"testing"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/manager"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/utils"
)

var (
	testServerAddress = "localhost"
	testServerPort    = "39783"
)

func TestAddReaper(t *testing.T) {

}

func TestUpdateReaper(t *testing.T) {

}

func TestDeleteReaper(t *testing.T) {

}

func TestListRunningReapers(t *testing.T) {

}

func TestStartManager(t *testing.T) {

}

func TestShutdownManager(t *testing.T) {

}

func setupGRPCServer() {
	server := utils.CreateServer(utils.DefaultHandler)
	testOptions := utils.GetTestOptions(server)
	manager.StartServer(testServerAddress, testServerPort, testOptions...)
}
