package main

import (
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/manager"
)

func main() {
	manager.StartServer("localhost", "8080")
}
