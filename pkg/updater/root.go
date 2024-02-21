package updater

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/icaruk/up-npm/pkg/utils/cli"
	npm "github.com/icaruk/up-npm/pkg/utils/npm"
	packagejson "github.com/icaruk/up-npm/pkg/utils/packagejson"
	repositorypkg "github.com/icaruk/up-npm/pkg/utils/repository"
	versionpkg "github.com/icaruk/up-npm/pkg/utils/version"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/logrusorgru/aurora/v4"
	"github.com/schollz/progressbar/v3"
	"github.com/tidwall/sjson"
)

type PackageJSON struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Homepage        string            `json:"homepage"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

func printUpdatablePackagesTable(versionComparison map[string]versionpkg.VersionComparisonItem) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Package", "Current", "Latest"})

	t.SetColumnConfigs(([]table.ColumnConfig{
		{
			Name:  "Package",
			Align: text.AlignLeft,
		},
		{
			Name:  "Current",
			Align: text.AlignRight,
		},
		{
			Name:  "Latest",
			Align: text.AlignRight,
		},
	}))

	// Add rows
	for key, value := range versionComparison {
		latestColorized := versionpkg.ColorizeVersion(value.Latest, value.VersionType)
		t.AppendRow(table.Row{key, value.Current, latestColorized})
	}

	t.Render()

}

func printSummary(totalCount int, majorCount int, minorCount int, patchCount int) {

	baseSt := fmt.Sprintf("Found %d packages to update", totalCount)

	var extraSt []string
	if patchCount > 0 {
		extraSt = append(
			extraSt,
			aurora.Sprintf(
				aurora.Green("%d patch"),
				patchCount,
			),
		)
	}
	if minorCount > 0 {
		extraSt = append(
			extraSt,
			aurora.Sprintf(
				aurora.Yellow("%d minor"),
				minorCount,
			),
		)
	}
	if majorCount > 0 {
		extraSt = append(
			extraSt,
			aurora.Sprintf(
				aurora.Red("%d major"),
				majorCount,
			),
		)
	}

	if len(extraSt) > 0 {
		joinedExtraSt := strings.Join(extraSt, ", ")
		baseSt = fmt.Sprintf("%s: %s", baseSt, joinedExtraSt)
	}

	fmt.Println(baseSt)

}

func countVersionTypes(
	versionComparison map[string]versionpkg.VersionComparisonItem,
) (
	majorCount int, minorCount int, patchCount int, totalCount int,
) {
	for _, value := range versionComparison {

		switch value.VersionType {
		case versionpkg.Major:
			majorCount++
		case versionpkg.Minor:
			minorCount++
		case versionpkg.Patch:
			patchCount++
		}

		totalCount++
	}

	return
}

type updatePackageOptions struct {
	update       string
	skip         string
	show_changes string
	finish       string
}

type writeJsonOptions struct {
	yes        string
	yes_backup string
	no         string
}

func promptWriteJson(options writeJsonOptions, file string) (string, error) {
	response := ""
	prompt := &survey.Select{
		Message: fmt.Sprintf("Update %s?", file),
		Options: []string{
			options.yes,
			options.yes_backup,
			options.no,
		},
	}
	err := survey.AskOne(prompt, &response)

	return response, err
}

type UpgradeType string
type UpgradeDirection string

const (
	UpgradeTypeNone  UpgradeType = "none"
	UpgradeTypeMajor UpgradeType = "major"
	UpgradeTypeMinor UpgradeType = "minor"
	UpgradeTypePatch UpgradeType = "patch"
)
const (
	UpgradeDirectionNone      UpgradeDirection = "none"
	UpgradeDirectionUpgrade   UpgradeDirection = "upgrade"
	UpgradeDirectionDowngrade UpgradeDirection = "downgrade"
)

func initProgressBar(maxBar int) *progressbar.ProgressBar {
	return progressbar.NewOptions(maxBar,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription("[cyan]Checking updates...[reset]"),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
}

func createPackageJsonBackup(file string) (bool, error) {
	date := time.Now().Format("2006-01-02-15-04-05")

	backupFileName := fmt.Sprintf("backup.%s.%s", date, file)

	fileData, err := os.ReadFile(file)
	if err != nil {
		fmt.Println(err)
		return false, err
	}

	err = os.WriteFile(backupFileName, fileData, 0644)
	if err != nil {
		fmt.Println(err)
		return false, err
	}

	return true, nil
}

// https://gist.github.com/hyg/9c4afcd91fe24316cbf0
func Openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

}

const concurrencyLimit int = 10

func readDependencies(
	dependencyList map[string]string,
	targetMap map[string]versionpkg.VersionComparisonItem,
	isDev bool,
	bar *progressbar.ProgressBar,
	filter string,
) (lockedDependencyCount int) {

	var wg sync.WaitGroup
	semaphoreChan := make(chan struct{}, concurrencyLimit)
	resultsChan := make(chan string, len(dependencyList))
	doneChan := make(chan struct{})

	defer func() {
		close(semaphoreChan)
	}()

	for packageName, currentVersion := range dependencyList {

		// Check filter
		if filter != "" {
			if !strings.Contains(packageName, filter) {
				continue
			}
		}

		// Get version and prefix
		versionPrefix, cleanCurrentVersion := versionpkg.GetCleanVersion(currentVersion)

		if cleanCurrentVersion == "" {
			fmt.Println(packageName, " has unsupported/invalid version, skipping...")
			continue
		}

		if versionPrefix == "" {
			fmt.Printf(" is locked to version %s, skipping...", cleanCurrentVersion)
			lockedDependencyCount++
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
			resp, err := npm.FetchNpmRegistry(dependency)
			if err != nil {
				fmt.Println("Failed to fetch", dependency, " from npm registry, skipping...")
				resultsChan <- "" // Enviar un resultado vacío para que se tenga en cuenta en el recuento de resultados
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				// Failed to fetch
				return
			}

			// Get response data
			var result map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&result)

			distTags := result["dist-tags"].(map[string]interface{})

			var homepage string
			if value, ok := result["homepage"]; ok {
				homepage = value.(string)
			}

			var repositoryUrl string
			if value, ok := result["repository"]; ok {
				repository := value.(map[string]interface{})["url"].(string)
				repositoryUrl = repositorypkg.GetRepositoryUrl(repository)
			}

			// Get latest version from distTags
			var latestVersion string
			for key := range distTags {
				if key == "latest" {
					latestVersion = distTags[key].(string)
				}
			}

			// Get version update type (major, minor, patch, none)
			upgradeType, upgradeDirection := versionpkg.GetVersionUpdateType(cleanCurrentVersion, latestVersion)

			// Save data
			if upgradeDirection == versionpkg.Upgrade {
				targetMap[dependency] = versionpkg.VersionComparisonItem{
					Current:       cleanCurrentVersion,
					Latest:        latestVersion,
					VersionType:   upgradeType,
					ShouldUpdate:  false,
					Homepage:      homepage,
					RepositoryUrl: repositoryUrl,
					VersionPrefix: versionPrefix,
					IsDev:         isDev,
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
	return lockedDependencyCount
}

func Init(cfg npm.CmdFlags, binVersion string) {

	var isFilterFilled bool = cfg.Filter != ""

	// New version message
	latestRelease, err := repositorypkg.FetchRepositoryLatestRelease("icaruk", "up-npm")

	if err == nil {

		latestReleaseVersion := latestRelease["tag_name"].(string)

		_, upgradeDirection := versionpkg.GetVersionUpdateType(binVersion, latestReleaseVersion)

		if upgradeDirection == versionpkg.UpgradeDirection(UpgradeDirectionUpgrade) {

			fmt.Println()

			fmt.Println(
				aurora.Sprintf(
					aurora.BrightGreen("Update: up-npm %s is available!"),
					aurora.Green(latestReleaseVersion),
				),
				aurora.Sprintf(
					aurora.Faint("(current version is %s)"),
					aurora.Faint(binVersion),
				),
			)
			// fmt.Println(aurora.Sprintf(aurora.Faint("Current version in %s"), aurora.Faint(binVersion)))
			fmt.Println(
				"Click",
				aurora.Blue("here").Hyperlink("https://github.com/Icaruk/up-npm/releases/latest"),
				"to check the latest changes.",
			)
		}

	}

	fmt.Println()

	dependencies, devDependencies, jsonFile, err := packagejson.GetDependenciesFromPackageJson(cfg.File, cfg.NoDev)

	if err != nil {
		fmt.Println(err)
		fmt.Println()
		return
	}

	versionComparison := map[string]versionpkg.VersionComparisonItem{}

	// Progress bar
	totalDependencyCount := len(dependencies) + len(devDependencies)
	bar := initProgressBar(totalDependencyCount)

	// Process dependencies
	var lockedDependencyCount int
	var lockedDevDependencyCount int

	lockedDependencyCount = readDependencies(dependencies, versionComparison, false, bar, cfg.Filter)

	// Process devDependencies
	if !cfg.NoDev {
		lockedDevDependencyCount = readDependencies(devDependencies, versionComparison, true, bar, cfg.Filter)
	}

	// Count total dependencies and filtered dependencies
	filteredDependencyCount := totalDependencyCount
	if isFilterFilled {
		filteredDependencyCount = len(versionComparison)
	}

	// Count version types
	majorCount, minorCount, patchCount, totalCount := countVersionTypes(versionComparison)
	if filteredDependencyCount == 0 {
		fmt.Println()
		fmt.Println()
		fmt.Println(aurora.Green("No outdated dependencies!"))
		fmt.Println()
		return
	}

	// Table
	fmt.Println("")
	fmt.Println("")
	printUpdatablePackagesTable(versionComparison)
	fmt.Println("")

	// Print summary line (1 major, 1 minor, 1 patch)
	if isFilterFilled {
		fmt.Println("Filtered", aurora.Blue(filteredDependencyCount), "dependencies from a total of", aurora.Blue(totalDependencyCount))
	} else {
		fmt.Println("Total dependencies: ", aurora.Cyan(filteredDependencyCount))

		totalLockedDependencyCount := lockedDependencyCount + lockedDevDependencyCount
		if totalLockedDependencyCount > 0 {
			s := fmt.Sprintf("Locked dependencies: %d", totalLockedDependencyCount)
			fmt.Println(aurora.Faint(s))
		}
	}

	printSummary(totalCount, majorCount, minorCount, patchCount)
	fmt.Println()

	// Prompt user to update each dependency
	updatePackageOptions := updatePackageOptions{
		update:       "Update",
		skip:         "Skip",
		show_changes: "Show changes",
		finish:       "Finish",
	}

	updateProgressCount := 1

	for key, value := range versionComparison {

		exit := false

		for {

			response := cli.PromptUpdateDependency(
				key,
				value.Current,
				value.Latest,
				value.VersionType,
			)

			updateProgressCount++

			if err != nil {
				if err == terminal.InterruptErr {
					log.Fatal("interrupted")
				}
			}

			if response == updatePackageOptions.skip {
				// Skipped dependencyName in green color
				fmt.Println(
					aurora.Sprintf(
						aurora.Faint("Skipped \"%s\""),
						key,
					),
				)

				break
			}

			if response == updatePackageOptions.show_changes {

				// Get user and repository from repository URL
				urlMetadata := repositorypkg.GetRepositoryUrlMetadata(value.RepositoryUrl)

				// Fetch repository from github
				_, err := repositorypkg.FetchRepositoryLatestRelease(urlMetadata.Username, urlMetadata.RepositoryName)
				missingLatestRelease := err != nil

				var changelogMdUrl string

				if missingLatestRelease {
					fmt.Println(aurora.Faint("Could not find latest release from github"))

					// Fetch CHANGELOG.md
					response, err := repositorypkg.FetchRepositoryChangelogFile(urlMetadata.Username, urlMetadata.RepositoryName)

					if err == nil {
						changelogMdUrl = response["html_url"].(string)
					} else {
						fmt.Println(aurora.Faint("Could not find CHANGELOG.md"))
					}
				}

				var url string

				if !missingLatestRelease {
					url = value.RepositoryUrl + "/releases" + "#:~:text=" + value.Current
				} else if changelogMdUrl != "" {
					url = changelogMdUrl
				} else {
					url = value.Homepage
				}

				if url == "" {
					fmt.Println(aurora.Yellow("No repository or homepage URL found"))
					break
				}

				fmt.Println("Opening...")
				fmt.Println()
				Openbrowser(url)

			}

			if response == updatePackageOptions.update {
				// get a copy of the entry
				if entry, ok := versionComparison[key]; ok {
					entry.ShouldUpdate = true      // then modify the copy
					versionComparison[key] = entry // then reassign map entry
				}

				colorizedVersion := versionpkg.ColorizeVersion(value.Latest, value.VersionType)

				fmt.Println(
					aurora.Sprintf(
						"%s \"%s\" from %s to %s",
						aurora.Green("Updated"),
						key,
						value.Current,
						colorizedVersion,
					),
				)

				break
			}

			if response == updatePackageOptions.finish {
				fmt.Println("Finished update process")
				exit = true
				break
			}
		}

		if exit {
			break
		}

	}

	fmt.Println()

	// ------------------------------------------

	// Check how many updates are on versionComparison with value.shouldUpdate = true
	var shouldUpdateCount int
	for _, value := range versionComparison {
		if value.ShouldUpdate {
			shouldUpdateCount++
		}
	}

	if shouldUpdateCount == 0 {
		fmt.Println(aurora.Yellow("No packages have been selected to update"))
		return
	}

	fmt.Println(
		aurora.Green("There are"),
		aurora.Green(shouldUpdateCount),
		aurora.Green("package(s) selected to be updated"),
	)
	fmt.Println()

	// Prompt to write package.json
	writeJsonOptions := writeJsonOptions{
		yes:        "Yes",
		yes_backup: "Yes, but backup before update",
		no:         "No",
	}

	response, err := promptWriteJson(writeJsonOptions, cfg.File)

	if err != nil {
		if err == terminal.InterruptErr {
			log.Fatal("interrupted")
		}
	}

	if response == writeJsonOptions.no {
		fmt.Println("Cancelled update process")
		return
	}

	if response == writeJsonOptions.yes_backup {
		createPackageJsonBackup(cfg.File)
	}

	// Stringify package.json
	jsonFileStr := string(jsonFile)

	// Write dependencies to package.json
	for key, value := range versionComparison {
		if value.ShouldUpdate {

			dependenciesKeyName := "dependencies"
			if value.IsDev {
				dependenciesKeyName = "devDependencies"
			}

			// If key includes a dor `.` replace with `\.`
			if strings.Contains(key, ".") {
				key = strings.ReplaceAll(key, ".", `\.`)
			}

			dotPath := fmt.Sprintf("%s.%s", dependenciesKeyName, key)
			latestVersion := fmt.Sprintf("%s%s", value.VersionPrefix, value.Latest)

			jsonFileStr, _ = sjson.Set(jsonFileStr, dotPath, latestVersion)
		}
	}

	// Write to file
	err = os.WriteFile(cfg.File, []byte(jsonFileStr), 0644)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println()

	fmt.Printf(
		"✅ %s has been updated with %s\n",
		cfg.File,
		aurora.Sprintf(
			aurora.Green("%d updated packages"),
			shouldUpdateCount,
		),
	)

	fmt.Println()

	fmt.Println("Run 'npm install' to install dependencies")
	fmt.Println()

}
