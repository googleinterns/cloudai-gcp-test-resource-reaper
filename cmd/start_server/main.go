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

package main

import (
	"context"
	"flag"
	"log"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/logger"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/manager"
)

func main() {
	port := flag.String("port", "8000", "port to run gRPC server on")
	projectID := flag.String("project-id", "", "GCP Project ID for where to store logs")
	logsName := flag.String("logs-name", "", "name of logs")

	flag.Parse()

	if err := logger.CreateLogger(); err != nil {
		log.Fatal(err)
	}
	defer logger.Close()
	if len(*projectID) > 0 && len(*logsName) > 0 {
		logger.AddCloudLogger(context.Background(), *projectID, *logsName)
		logger.Logf("Logging to %s in project", *logsName, *projectID)
	}

	manager.StartServer(*port)
}
