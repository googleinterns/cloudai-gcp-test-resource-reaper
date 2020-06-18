package reaper

import (
	"context"
	"fmt"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
)

func setupResources() {

}

func Example() {
	reaperConfig := &reaperconfig.ReaperConfig{
		Resources:  []*reaperconfig.ResourceConfig{},
		Schedule:   "* * * * *",
		SkipFilter: "",
		ProjectId:  "",
		Uuid:       "",
	}
	context := context.Background()

	reaper := NewReaper(context, reaperConfig)
	fmt.Println("Running Example Func")
	// Output: Running Example Func
}
