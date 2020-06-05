package main

import (
	"fmt"
	"log"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/pkg/clients"
	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
)

func main() {
	var c clients.Client
	c = &clients.GCEClient{}
	err := c.Auth()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully Authed")
	testConfig := reaperconfig.ResourceConfig{
		ResourceType: reaperconfig.ResourceType_GCE_VM,
		NameFilter:   "",
		SkipFilter:   "",
		Zones:        []string{"us-central1-a", "us-east4-c"},
		Ttl:          "",
	}
	res, err := c.GetResources("supercclank-test-df-staging", testConfig)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res)

	// testInstance := res[len(res)-1]
	// c.DeleteResource("supercclank-test-df-staging", testInstance)

	// res, err = c.GetResources("supercclank-test-df-staging", testConfig)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(res)
}
