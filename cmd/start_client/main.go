package main

import (
	"flag"
)

// Binary with flags
// https://golang.org/pkg/flag/

func main() {
	createReaper := flag.NewFlagSet("create", flag.ExitOnError)

	flag.Parse()

	// if *createReaper {
	// 	fmt.Println("True")
	// } else {
	// 	fmt.Println("FALSE")
	// }
	// reaperClient := client.StartClient(context.Background(), "localhost", "8000")
	// defer reaperClient.Close()

	// reaperClient.ShutdownManager()
	// reaperClient.StartManager()
	// reaperClient.ListRunningReapers()
}
