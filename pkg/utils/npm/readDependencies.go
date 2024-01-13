package npm

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	repositorypkg "github.com/icaruk/up-npm/pkg/utils/repository"
	"github.com/icaruk/up-npm/pkg/utils/version"

	"github.com/schollz/progressbar/v3"
)

type CmdFlags struct {
	NoDev          bool
	AllowDowngrade bool
	Filter         string
}

type VersionComparisonItem struct {
	current       string
	latest        string
	versionType   version.UpgradeType
	shouldUpdate  bool
	homepage      string
	repositoryUrl string
	versionPrefix string
	isDev         bool
}

const concurrencyLimit int = 10

func ReadDependencies(
	dependencyList map[string]string,
	targetMap map[string]VersionComparisonItem,
	isDev bool,
	bar *progressbar.ProgressBar,
	cfg CmdFlags,
) {

	var wg sync.WaitGroup
	semaphoreChan := make(chan struct{}, concurrencyLimit)
	resultsChan := make(chan string, len(dependencyList))
	doneChan := make(chan struct{})

	defer func() {
		close(semaphoreChan)
	}()

	for packageName, currentVersion := range dependencyList {

		// Check filter
		if cfg.Filter != "" {
			if !strings.Contains(packageName, cfg.Filter) {
				continue
			}
		}

		// Get version and prefix
		versionPrefix, cleanCurrentVersion := version.GetCleanVersion(currentVersion)

		if cleanCurrentVersion == "" {
			fmt.Println(packageName, " has unsupported/invalid version, skipping...")
			continue
		}

		wg.Add(1)

		go func(dependency string, currentVersion string) {
			defer func() {
				wg.Done()
				<-semaphoreChan
			}()
			semaphoreChan <- struct{}{}

			// Perform get request to npm registry
			resp, err := FetchNpmRegistry(dependency)
			if err != nil {
				fmt.Println("Failed to fetch", dependency, " from npm registry, skipping...")
				resultsChan <- "" // Enviar un resultado vacío para que se tenga en cuenta en la cuenta de resultados
				return
			}
			defer resp.Body.Close()

			// Get response data
			var result map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&result)

			distTags := result["dist-tags"].(map[string]interface{})

			var homepage string
			if result["homepage"] != nil {
				homepage = result["homepage"].(string)
			}

			var repositoryData map[string]interface{}
			if result["repository"] != nil {
				repositoryData = result["repository"].(map[string]interface{})
			}

			var repositoryUrl string
			if repositoryData["url"] != nil {
				repositoryUrl = repositorypkg.GetRepositoryUrl(repositoryData["url"].(string))
			}

			// Get latest version from distTags
			var latestVersion string
			for key := range distTags {
				if key == "latest" {
					latestVersion = distTags[key].(string)
				}
			}

			// Get version update type (major, minor, patch, none)
			upgradeType, upgradeDirection := version.GetVersionUpdateType(cleanCurrentVersion, latestVersion)

			// Save data
			if (upgradeDirection == version.Upgrade) ||
				(cfg.AllowDowngrade && upgradeDirection == version.Downgrade) {
				targetMap[dependency] = VersionComparisonItem{
					current:       cleanCurrentVersion,
					latest:        latestVersion,
					versionType:   upgradeType,
					shouldUpdate:  false,
					homepage:      homepage,
					repositoryUrl: repositoryUrl,
					versionPrefix: versionPrefix,
					isDev:         isDev,
				}
			}

			resultsChan <- ""

			if bar != nil {
				bar.Add(1)
			}

		}(packageName, currentVersion)

	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(doneChan)

}
