package resources

import (
	"testing"
	"time"

	"github.com/googleinterns/cloudai-gcp-test-resource-reaper/reaperconfig"
)

type ShouldWatchTestCase struct {
	TestResource Resource
	NameFilter   string
	SkipFilter   string
	Expected     bool
}

// Variables whose values are not important for testing purposes
var (
	zone              = "testZone"
	timeCreatedString = "2019-10-12T07:20:50.52Z"
	timeCreated, _    = time.Parse(time.RFC3339, timeCreatedString)
	resourceType      = reaperconfig.ResourceType_GCE_VM
)

var testShouldWatchCases = []ShouldWatchTestCase{
	ShouldWatchTestCase{
		NewResource("testName", zone, timeCreated, resourceType),
		"test", "", true,
	},
	ShouldWatchTestCase{
		NewResource("testName", zone, timeCreated, resourceType),
		"test", "test", false,
	},
	ShouldWatchTestCase{
		NewResource("testName", zone, timeCreated, resourceType),
		"another|test", "", true,
	},
	ShouldWatchTestCase{
		NewResource("testName", zone, timeCreated, resourceType),
		"another|test", "test", false,
	},
	ShouldWatchTestCase{
		NewResource("testName", zone, timeCreated, resourceType),
		"another", "", false,
	},
	ShouldWatchTestCase{
		NewResource("testName", zone, timeCreated, resourceType),
		"another", "test", false,
	},
	ShouldWatchTestCase{
		NewResource("testName", zone, timeCreated, resourceType),
		"", "test", false,
	},
	ShouldWatchTestCase{
		NewResource("testName", zone, timeCreated, resourceType),
		"", "", false,
	},
}

func TestShouldAddResourceToWatchlist(t *testing.T) {
	for i, testCase := range testShouldWatchCases {
		result := ShouldAddResourceToWatchlist(
			testCase.TestResource, testCase.NameFilter, testCase.SkipFilter,
		)
		if result != testCase.Expected {
			t.Errorf("Expected %t, got %t", testCase.Expected, result, i)
		}
	}
}
