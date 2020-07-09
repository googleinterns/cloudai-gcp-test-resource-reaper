package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/client"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/reaper"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
)

func main() {
	deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
	deleteUUID := deleteCmd.String("uuid", "", "UUID of the reaper")

	if len(os.Args) < 2 {
		fmt.Println("expected 'create', 'update', 'list', 'delete', 'start', or 'shutdown' commands")
		os.Exit(1)
	}

	reaperClient := client.StartClient(context.Background(), "localhost", "8000")
	defer reaperClient.Close()

	switch os.Args[1] {
	case "create":
		config, err := createReaperConfigPrompt()
		if err != nil {
			fmt.Println("Creating reaper config failed with the following error: ", err.Error())
			os.Exit(1)
		}
		uuid, err := reaperClient.AddReaper(config)
		if err != nil {
			fmt.Println("Create reaper failed with following error: ", err.Error())
			os.Exit(1)
		}
		fmt.Printf("Reaper with UUID %s successfully created\n", uuid)

	case "update":
		config, err := createReaperConfigPrompt()
		if err != nil {
			fmt.Println("Creating reaper config failed with the following error: ", err.Error())
			os.Exit(1)
		}
		uuid, err := reaperClient.UpdateReaper(config)
		if err != nil {
			fmt.Println("Create reaper failed with following error: ", err.Error())
			os.Exit(1)
		}
		fmt.Printf("Reaper with UUID %s successfully updated\n", uuid)

	case "list":
		reapers, err := reaperClient.ListRunningReapers()
		if err != nil {
			fmt.Println("List reapers failed with following error: ", err.Error())
			os.Exit(1)
		}
		fmt.Println("Running Reaper UUIDs: ", strings.Join(reapers, ", "))

	case "delete":
		deleteCmd.Parse(os.Args[2:])
		if len(strings.TrimSpace(*deleteUUID)) == 0 {
			fmt.Println("uuid of reaper to delete cannot be empty")
			os.Exit(1)
		}
		err := reaperClient.DeleteReaper(*deleteUUID)
		if err != nil {
			fmt.Println("Delete reapers failed with following error: ", err.Error())
			os.Exit(1)
		}
		fmt.Printf("Reaper with UUID %s successfully deleted\n", *deleteUUID)

	case "start":
		err := reaperClient.StartManager()
		if err != nil {
			fmt.Println("Start manager failed with following error: ", err.Error())
		}
		fmt.Println("Reaper manager started")

	case "shutdown":
		err := reaperClient.ShutdownManager()
		if err != nil {
			fmt.Println("Shutdown manager failed with following error: ", err.Error())
		}
		fmt.Println("Reaper manager shutdown")

	default:
		fmt.Println("expected 'create', 'update', 'list', 'delete', 'start', or 'shutdown' commands")
		os.Exit(1)
	}
}

func createReaperConfigPrompt() (*reaperconfig.ReaperConfig, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Reaper Configuration Setup")

	fmt.Print("Reaper UUID: ")
	uuid, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	uuid = strings.TrimSpace(uuid)

	fmt.Print("Project ID: ")
	projectID, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	projectID = strings.TrimSpace(projectID)

	fmt.Print("Reaper run schedule (in cron time string format): ")
	schedule, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	schedule = strings.TrimSpace(schedule)

	var resources []*reaperconfig.ResourceConfig
	for {
		fmt.Print("Add another resource? (y/n): ")
		response, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if len(response) == 1 || response[0] == 'n' {
			break
		}

		fmt.Print("Resource type: ")
		resourceTypeString, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		resourceTypeString = strings.TrimSpace(resourceTypeString)
		var resourceType reaperconfig.ResourceType

		switch resourceTypeString {
		case "GCE_VM":
			resourceType = reaperconfig.ResourceType_GCE_VM
		case "GCS_Bucket":
			resourceType = reaperconfig.ResourceType_GCS_BUCKET
		case "GCS_Object":
			resourceType = reaperconfig.ResourceType_GCS_OBJECT
		default:
			return nil, fmt.Errorf("Invalid resource type %s", resourceTypeString)
		}

		fmt.Print("Zones (comma separated list): ")
		zonesString, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		zonesString = strings.TrimSpace(zonesString)
		zones := strings.Split(zonesString, ",")

		fmt.Print("Name filter: ")
		nameFilter, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		nameFilter = strings.TrimSpace(nameFilter)

		fmt.Print("Skip filter: ")
		skipFilter, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		skipFilter = strings.TrimSpace(skipFilter)

		fmt.Print("TTL (in cron time string format): ")
		ttl, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		ttl = strings.TrimSpace(ttl)

		resources = append(
			resources,
			reaper.NewResourceConfig(resourceType, zones, nameFilter, skipFilter, ttl),
		)
	}

	config := reaper.NewReaperConfig(resources, schedule, projectID, uuid)
	return config, nil
}
