package updater

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	npm "github.com/icaruk/up-npm/pkg/utils/npm"
	packagejson "github.com/icaruk/up-npm/pkg/utils/packagejson"
	repositorypkg "github.com/icaruk/up-npm/pkg/utils/repository"
	"github.com/icaruk/up-npm/pkg/utils/version"
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

type NpmRegistryPackage struct {
	ID       string `json:"_id"`
	Rev      string `json:"_rev"`
	Name     string `json:"name"`
	DistTags struct {
		Latest string `json:"latest"`
	} `json:"dist-tags"`
	Versions struct {
		Version map[string]interface{}
	} `json:"versions"`
	Homepage   string `json:"homepage"`
	Repository struct {
		Type string `json:"type"`
		URL  string `json:"url"`
	} `json:"repository"`
}

func colorizeVersion(version string, versionType versionpkg.UpgradeType) string {
	vLatestMajor, vLatestMinor, vLatestPatch := getVersionComponents(version)

	var colorizedVersion string

	if versionType == versionpkg.Major {
		colorizedVersion = aurora.Sprintf(aurora.Red("%d.%d.%d"), aurora.Red(vLatestMajor), aurora.Red(vLatestMinor), aurora.Red(vLatestPatch))
	} else if versionType == versionpkg.Minor {
		colorizedVersion = aurora.Sprintf(aurora.White("%d.%d.%d"), aurora.White(vLatestMajor), aurora.Yellow(vLatestMinor), aurora.Yellow(vLatestPatch))
	} else if versionType == versionpkg.Patch {
		colorizedVersion = aurora.Sprintf(aurora.White("%d.%d.%d"), aurora.White(vLatestMajor), aurora.White(vLatestMinor), aurora.Green(vLatestPatch))
	}

	return colorizedVersion
}

func printUpdateProgress(current int, max int) string {
	return fmt.Sprintf(
		"(%d/%d)",
		current,
		max,
	)
}

func printUpdatablePackagesTable(versionComparison map[string]VersionComparisonItem) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Package", "Current", "Latest"})

	// Add rows
	for key, value := range versionComparison {
		latestColorized := colorizeVersion(value.latest, value.versionType)
		t.AppendRow(table.Row{key, value.current, latestColorized}, table.RowConfig{AutoMergeAlign: text.AlignRight})
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

func getVersionComponents(semver string) (int, int, int) {
	re := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)
	matches := re.FindStringSubmatch(semver)
	vCurrentMajor, _ := strconv.Atoi(matches[1])
	vCurrentMinor, _ := strconv.Atoi(matches[2])
	vCurrentPatch, _ := strconv.Atoi(matches[3])

	return vCurrentMajor, vCurrentMinor, vCurrentPatch
}

func countVersionTypes(
	versionComparison map[string]VersionComparisonItem,
) (
	majorCount int, minorCount int, patchCount int, totalCount int,
) {
	for _, value := range versionComparison {

		fmt.Println(value)

		switch value.versionType {
		case version.Major:
			majorCount++
		case version.Minor:
			minorCount++
		case version.Patch:
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

func promptUpdateDependency(
	options updatePackageOptions,
	dependency string,
	currentVersion string,
	latestVersion string,
	versionType version.UpgradeType,
	updateProgressCount int,
	maxUpdateProgress int,
	isDevDependency bool,
) (string, error) {

	isDevDependencyText := ""
	if isDevDependency {
		isDevDependencyText = "(devDependency)"
	}

	response := ""
	prompt := &survey.Select{
		Message: fmt.Sprintf(
			"%s Update \"%s\" %s from %s to %s?",
			printUpdateProgress(updateProgressCount, maxUpdateProgress),
			dependency,
			isDevDependencyText,
			currentVersion,
			colorizeVersion(latestVersion, versionType),
		),
		Help: "Use arrow keys to navigate",
		Options: []string{
			options.update,
			options.skip,
			options.show_changes,
			options.finish,
		},
	}
	err := survey.AskOne(prompt, &response)

	return response, err
}

func promptWriteJson(options writeJsonOptions) (string, error) {
	response := ""
	prompt := &survey.Select{
		Message: "Update package.json?",
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

func createPackageJsonBackup() (bool, error) {
	date := time.Now().Format("2006-01-02-15-04-05")

	backupFileName := fmt.Sprintf("backup.%s.package.json", date)

	file, err := os.ReadFile("package.json")
	if err != nil {
		fmt.Println(err)
		return false, err
	}

	err = os.WriteFile(backupFileName, file, 0644)
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
	targetMap map[string]VersionComparisonItem,
	isDev bool,
	bar *progressbar.ProgressBar,
	filter string,
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
		if filter != "" {
			if !strings.Contains(packageName, filter) {
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
			resp, err := npm.FetchNpmRegistry(dependency)
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

			var repository map[string]interface{}
			if result["repository"] != nil {
				repository = result["repository"].(map[string]interface{})
			}

			var repositoryUrl string
			if repository["url"] != nil {
				repositoryUrl = repositorypkg.GetRepositoryUrl(repository["url"].(string))
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
			if upgradeDirection == version.Upgrade {
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

func Init(cfg npm.CmdFlags) {

	var isFilterFilled bool = cfg.Filter != ""

	fmt.Println()

	dependencies, devDependencies, jsonFile, err := packagejson.GetDependenciesFromPackageJson("package.json", cfg.NoDev)

	if err != nil {
		fmt.Println(err)
		fmt.Println()
		return
	}

	versionComparison := map[string]VersionComparisonItem{}

	// Progress bar
	maxBar := len(dependencies)
	bar := initProgressBar(maxBar)

	// Process dependencies
	dependencyCount := len(dependencies)
	readDependencies(dependencies, versionComparison, false, bar, cfg.Filter)

	if cfg.NoDev == false {
		dependencyCount += len(devDependencies)
		readDependencies(devDependencies, versionComparison, true, bar, cfg.Filter)
	}

	// Count total dependencies and filtered dependencies
	filteredDependencyCount := dependencyCount
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
		fmt.Println("Filtered", aurora.Blue(filteredDependencyCount), "dependencies from a total of", aurora.Blue(dependencyCount))
	} else {
		fmt.Println("Total dependencies: ", aurora.Cyan(filteredDependencyCount))
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

			response, err := promptUpdateDependency(
				updatePackageOptions,
				key,
				value.current,
				value.latest,
				value.versionType,
				updateProgressCount,
				totalCount,
				value.isDev,
			)
			updateProgressCount++

			if err != nil {
				if err == terminal.InterruptErr {
					log.Fatal("interrupted")
				}
			}

			if response == updatePackageOptions.skip {
				break
			}

			if response == updatePackageOptions.show_changes {
				// Open browser url
				var url string

				if value.repositoryUrl != "" {
					url = value.repositoryUrl + "/releases" + "#:~:text=" + value.current
				} else {
					url = value.homepage
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
					entry.shouldUpdate = true      // then modify the copy
					versionComparison[key] = entry // then reassign map entry
				}
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
		if value.shouldUpdate {
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

	response, err := promptWriteJson(writeJsonOptions)

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
		createPackageJsonBackup()
	}

	// Stringify package.json
	jsonFileStr := string(jsonFile)

	// Write dependencies to package.json
	for key, value := range versionComparison {
		if value.shouldUpdate {

			dependenciesKeyName := "dependencies"
			if value.isDev {
				dependenciesKeyName = "devDependencies"
			}

			dotPath := fmt.Sprintf("%s.%s", dependenciesKeyName, key)
			latestVersion := fmt.Sprintf("%s%s", value.versionPrefix, value.latest)

			jsonFileStr, _ = sjson.Set(jsonFileStr, dotPath, latestVersion)
		}
	}

	// Write to file
	err = os.WriteFile("package.json", []byte(jsonFileStr), 0644)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println()
	fmt.Println(
		"✅ package.json has been updated with",
		aurora.Sprintf(aurora.Green("%d updated packages"), shouldUpdateCount),
	)
	fmt.Println()

	fmt.Println("Run 'npm install' to install dependencies")
	fmt.Println()

}
