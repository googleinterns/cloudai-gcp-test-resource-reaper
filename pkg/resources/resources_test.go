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
	for _, testCase := range testShouldWatchCases {
		result := ShouldAddResourceToWatchlist(
			testCase.TestResource, testCase.NameFilter, testCase.SkipFilter,
		)
		if result != testCase.Expected {
			t.Errorf("Expected %t, got %t", testCase.Expected, result)
		}
	}
}
