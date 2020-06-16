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

// Variables whose values are not important for testing purposes
var (
	zone         = "testZone"
	earlyTime    = time.Now().AddDate(-10, 0, 0)
	lateTime     = time.Now().AddDate(10, 0, 0)
	minuteAgo    = time.Now().Add(-1 * time.Minute)
	minuteLater  = time.Now().Add(time.Minute)
	resourceType = reaperconfig.ResourceType_GCE_VM
)

type ShouldWatchTestCase struct {
	TestResource Resource
	NameFilter   string
	SkipFilter   string
	Expected     bool
}

var testShouldWatchCases = []ShouldWatchTestCase{
	ShouldWatchTestCase{
		NewResource("testName", zone, time.Now(), resourceType),
		"test", "", true,
	},
	ShouldWatchTestCase{
		NewResource("testName", zone, time.Now(), resourceType),
		"test", "test", false,
	},
	ShouldWatchTestCase{
		NewResource("testName", zone, time.Now(), resourceType),
		"another|test", "", true,
	},
	ShouldWatchTestCase{
		NewResource("testName", zone, time.Now(), resourceType),
		"another|test", "test", false,
	},
	ShouldWatchTestCase{
		NewResource("testName", zone, time.Now(), resourceType),
		"another", "", false,
	},
	ShouldWatchTestCase{
		NewResource("testName", zone, time.Now(), resourceType),
		"another", "test", false,
	},
	ShouldWatchTestCase{
		NewResource("testName", zone, time.Now(), resourceType),
		"", "test", false,
	},
	ShouldWatchTestCase{
		NewResource("testName", zone, time.Now(), resourceType),
		"", "", false,
	},
}

// TestShouldAddResourceToWatchlist tests the ShouldAddResourceToWatchlist funcion.
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

type ReadyForDeletionTestCase struct {
	TestResource WatchedResource
	Expected     bool
}

var readyForDeletionTestCases = []ReadyForDeletionTestCase{
	ReadyForDeletionTestCase{createTestWatchedResource(earlyTime, "* * * * *"), true},
	ReadyForDeletionTestCase{createTestWatchedResource(earlyTime, "20 * 3 10 *"), true},
	ReadyForDeletionTestCase{createTestWatchedResource(minuteAgo, "* * * * *"), true},
	ReadyForDeletionTestCase{createTestWatchedResource(minuteAgo, "2 * * * *"), false},
	ReadyForDeletionTestCase{createTestWatchedResource(minuteLater, "* * * * *"), false},
	ReadyForDeletionTestCase{createTestWatchedResource(minuteLater, "10 * * * *"), false},
	ReadyForDeletionTestCase{createTestWatchedResource(lateTime, "* * * * *"), false},
	ReadyForDeletionTestCase{createTestWatchedResource(lateTime, "* 5 * * 7"), false},
}

func TestIsReadyForDeletion(t *testing.T) {
	for _, testCase := range readyForDeletionTestCases {
		result := testCase.TestResource.IsReadyForDeletion()
		if result != testCase.Expected {
			t.Errorf("Expected %t, got %t", testCase.Expected, result)
		}
	}
}

func createTestWatchedResource(creationTime time.Time, ttl string) WatchedResource {
	return WatchedResource{
		NewResource("TestResource", zone, creationTime, resourceType),
		ttl,
	}
}
