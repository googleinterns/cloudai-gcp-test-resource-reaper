package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/client"
)

func main() {
	deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
	deleteUuid := deleteCmd.String("uuid", "", "UUID of the reaper")

	if len(os.Args) < 2 {
		fmt.Println("expected 'create', 'update', 'list', 'delete', 'start', or 'shutdown' commands")
		os.Exit(1)
	}

	reaperClient := client.StartClient(context.Background(), "localhost", "8000")
	defer reaperClient.Close()

	switch os.Args[1] {
	case "create":
		createReaperConfigPrompt()
	case "update":
		createReaperConfigPrompt()
	case "list":
		reapers, err := reaperClient.ListRunningReapers()
		if err != nil {
			fmt.Println("List reapers failed with following error: ", err.Error())
			os.Exit(1)
		}
		fmt.Println("Running Reaper UUIDs: ", strings.Join(reapers, ", "))
	case "delete":
		deleteCmd.Parse(os.Args[2:])
		err := reaperClient.DeleteReaper(*deleteUuid)
		if err != nil {
			fmt.Println("Delete reapers failed with following error: ", err.Error())
			os.Exit(1)
		}
		fmt.Printf("Reaper with UUID %s successfully deleted", *deleteUuid)
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

func createReaperConfigPrompt() {

}
