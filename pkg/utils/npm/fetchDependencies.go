package npm

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	repositorypkg "github.com/icaruk/up-npm/pkg/utils/repository"
	"github.com/icaruk/up-npm/pkg/utils/version"

	"github.com/schollz/progressbar/v3"
)

type CmdFlags struct {
	NoDev          bool
	AllowDowngrade bool
	Filter         string
	File           string
	UpdatePatches  bool
}

const concurrencyLimit int = 10

func FetchDependencies(
	dependencyList map[string]string,
	targetMap map[string]version.VersionComparisonItem,
	isDev bool,
	token string,
	bar *progressbar.ProgressBar,
	cfg CmdFlags,
) (lockedDependencyCount int) {

	var wg sync.WaitGroup
	var mutex sync.Mutex // to protect targetMap access
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
			body, err := FetchNpmRegistry(dependency, token)
			if err != nil {
				fmt.Println("Failed to fetch", dependency, "from npm registry, skipping...")
				resultsChan <- "" // Enviar un resultado vacío para que se tenga en cuenta en la cuenta de resultados
				return
			}

			distTags := body["dist-tags"].(map[string]any)

			var homepage string
			if body["homepage"] != nil {
				homepage = body["homepage"].(string)
			}

			var repositoryData map[string]any
			if body["repository"] != nil {
				repositoryData = body["repository"].(map[string]any)
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

			// Get latest version inside time
			versionTimes := body["time"].(map[string]any)
			var latestReleaseDate string

			for key := range versionTimes {
				if key == latestVersion {
					latestReleaseDate = versionTimes[key].(string)
				}
			}

			hoursSinceLasRelease := -1.0

			// Parse ISOstring to date
			latestReleaseDateParsed, err := time.Parse(time.RFC3339, latestReleaseDate)
			if err != nil {
				fmt.Println("Failed to parse latest release date for", dependency, "skipping...")
			}

			// Get difference in hours
			hoursSinceLasRelease = time.Since(latestReleaseDateParsed).Hours()

			// Round to 2 decimals
			hoursSinceLasRelease = math.Round(hoursSinceLasRelease*10) / 10

			// Get version update type (major, minor, patch, none)
			upgradeType, upgradeDirection := version.GetVersionUpdateType(cleanCurrentVersion, latestVersion)

			// Save data
			if (upgradeDirection == version.Upgrade) ||
				(cfg.AllowDowngrade && upgradeDirection == version.Downgrade) {
				mutex.Lock()
				targetMap[dependency] = version.VersionComparisonItem{
					Current:              cleanCurrentVersion,
					Latest:               latestVersion,
					VersionType:          upgradeType,
					ShouldUpdate:         false,
					Homepage:             homepage,
					RepositoryUrl:        repositoryUrl,
					VersionPrefix:        versionPrefix,
					IsDev:                isDev,
					HoursSinceLasRelease: hoursSinceLasRelease,
				}
				mutex.Unlock()
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

	return lockedDependencyCount

}
