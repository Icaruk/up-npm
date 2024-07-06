package packagejson

import (
	"os"
	"sync"
)

const PACKAGE_MANAGERS_COUNT = 4

func GetPackageManager() string {

	var wg sync.WaitGroup
	results := make(chan string, PACKAGE_MANAGERS_COUNT)

	checkFile := func(filename, manager string) {
		defer wg.Done()
		if _, err := os.Stat(filename); err == nil {
			results <- manager
		}
	}

	wg.Add(PACKAGE_MANAGERS_COUNT)
	go checkFile("package-lock.json", "npm")
	go checkFile("pnpm-lock.yaml", "pnpm")
	go checkFile("bun.lockb", "bun")
	go checkFile("yarn.lock", "yarn")

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		return result
	}

	return "npm"

}

func GetInstallationCommand(packageManager string) string {
	switch packageManager {
	case "npm":
		return "npm install"
	case "pnpm":
		return "pnpm install"
	case "yarn":
		return "yarn install"
	case "bun":
		return "bun install"
	default:
		return "npm install"
	}
}
