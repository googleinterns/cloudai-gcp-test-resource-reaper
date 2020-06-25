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

package integration_tests

import (
	"encoding/json"
	"log"
	"os"
)

// TestConfig is a helper struct for reading from the config.json file
// for configuring intergation tests.
type TestConfig struct {
	ProjectID   string `json:"projectId"`
	AccessToken string `json:"accessToken"`
}

// ReadConfigFile reads the config.json file, and returns the ProjectID and accessToken
// specified in the file.
func ReadConfigFile() (string, string) {
	var configData TestConfig
	jsonConfigFile, err := os.Open("config.json")
	defer jsonConfigFile.Close()
	if err != nil {
		log.Fatal(err)
	}
	configParser := json.NewDecoder(jsonConfigFile)
	configParser.Decode(&configData)

	projectID := configData.ProjectID
	accessToken := configData.AccessToken
	return projectID, accessToken
}
