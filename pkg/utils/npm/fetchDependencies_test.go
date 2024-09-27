package npm_test

import (
	"sync"
	"testing"

	"github.com/icaruk/up-npm/pkg/utils/npm"
	"github.com/icaruk/up-npm/pkg/utils/version"
	"github.com/schollz/progressbar/v3"
)

func TestFetchDependencies(t *testing.T) {
	dependencyList := map[string]string{
		"@nestjs/core": "^10.3.9",
		"ajv":          "^8.16.0",
		"axios":        "^1.7.2",
		"bson":         "^6.7.0",
		"fastify":      "^4.27.0",
		"@tiptap/core": "^2.2.1",
	}

	// target map to be populated by FetchDependencies
	targetMap := make(map[string]version.VersionComparisonItem)
	var wg sync.WaitGroup
	bar := progressbar.New(3)
	token := "dummy-token"

	cfg := npm.CmdFlags{
		NoDev:          false,
		AllowDowngrade: true,
		Filter:         "",
		File:           "",
		UpdatePatches:  false,
	}

	// Simulate N concurrent goroutines
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			npm.FetchDependencies(dependencyList, targetMap, false, token, bar, cfg)
		}(i)
	}

	wg.Wait()

	// Verify that no concurrency issues occurred and check the map content
	t.Logf("Final targetMap size: %d", len(targetMap))
	if len(targetMap) != len(dependencyList) {
		t.Errorf("Expected %d dependencies in targetMap, got %d", len(dependencyList), len(targetMap))

		t.Logf("Dependency list:\n")

		for k, v := range dependencyList {
			t.Logf("%v: %v", k, v)
		}

		t.Logf("Target map:\n")

		for k, v := range targetMap {
			t.Logf("%v: %v", k, v)
		}
	}
}
