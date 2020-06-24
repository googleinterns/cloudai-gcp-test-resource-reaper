package integration_tests

import (
	"encoding/json"
	"log"
	"os"
)

type TestConfig struct {
	ProjectID   string `json:"projectId"`
	AccessToken string `json:"accessToken"`
}

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
