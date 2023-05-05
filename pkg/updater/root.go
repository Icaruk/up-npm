package updater

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/logrusorgru/aurora/v4"
	"github.com/schollz/progressbar/v3"
	"github.com/tidwall/sjson"
)

type CmdFlags struct {
	Dev            bool
	AllowDowngrade bool
	Filter         string
}

type PackageJSON struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Homepage        string            `json:"homepage"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
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

type VersionComparisonItem struct {
	current       string
	latest        string
	versionType   string
	shouldUpdate  bool
	homepage      string
	repositoryUrl string
	versionPrefix string
	isDev         bool
}

func colorizeVersion(version string, versionType string) string {
	vLatestMajor, vLatestMinor, vLatestPatch := getVersionComponents(version)

	var colorizedVersion string

	if versionType == "major" {
		colorizedVersion = aurora.Sprintf(aurora.Red("%d.%d.%d"), aurora.Red(vLatestMajor), aurora.Red(vLatestMinor), aurora.Red(vLatestPatch))
	} else if versionType == "minor" {
		colorizedVersion = aurora.Sprintf(aurora.White("%d.%d.%d"), aurora.White(vLatestMajor), aurora.Yellow(vLatestMinor), aurora.Yellow(vLatestPatch))
	} else if versionType == "patch" {
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
	// t.AppendFooter(table.Row{"", "Total", "1234"})

	// Add rows
	for key, value := range versionComparison {
		latestColorized := colorizeVersion(value.latest, value.versionType)
		t.AppendRow(table.Row{key, value.current, latestColorized})
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

func fetchNpmRegistry(dependency string) (*http.Response, error) {
	resp, err := http.Get(fmt.Sprintf("https://registry.npmjs.org/%s", dependency))
	if err != nil {
		fmt.Println(err)
	}

	return resp, err
}

func countVersionTypes(
	versionComparison map[string]VersionComparisonItem,
) (
	majorCount int, minorCount int, patchCount int, totalCount int,
) {
	for _, value := range versionComparison {

		fmt.Println(value)

		switch value.versionType {
		case "major":
			majorCount++
		case "minor":
			minorCount++
		case "patch":
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
	dependency,
	currentVersion,
	latestVersion,
	versionType string,
	updateProgressCount int,
	maxUpdateProgress int,
) (string, error) {
	response := ""
	prompt := &survey.Select{
		Message: fmt.Sprintf(
			"%s Update \"%s\" from %s to %s?",
			printUpdateProgress(updateProgressCount, maxUpdateProgress),
			dependency,
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
type UpgradeDirection int

const (
	UpgradeTypeNone  UpgradeType = "none"
	UpgradeTypeMajor UpgradeType = "major"
	UpgradeTypeMinor UpgradeType = "minor"
	UpgradeTypePatch UpgradeType = "patch"
)
const (
	UpgradeDirectionNone      UpgradeDirection = 0
	UpgradeDirectionUpgrade   UpgradeDirection = 1
	UpgradeDirectionDowngrade UpgradeDirection = -1
)

func getVersionUpdateType(currentVersion, latestVersion string) (upgradeType string, upgradeDirection int) {

	if latestVersion == currentVersion {
		return "none", 0
	}

	currentArr := strings.Split(currentVersion, ".") // 1.0.0
	latestArr := strings.Split(latestVersion, ".")   // 0.9.9

	versionChange := []int{0, 0, 0}

	for i := 0; i < len(currentArr); i++ {
		vCur, _ := strconv.Atoi(currentArr[i])
		vLat, _ := strconv.Atoi(latestArr[i])

		// 0 = major
		// 1 = minor
		// 2 = patch

		var changeDirection int

		if vLat > vCur {
			changeDirection = 1
		} else if vLat < vCur {
			changeDirection = -1
		}

		versionChange[i] = changeDirection

	}

	for i, change := range versionChange {
		if change == -1 {
			if i == 0 {
				return "major", change
			} else if i == 1 {
				return "minor", change
			} else {
				return "patch", change
			}
		}
		if change == 1 {
			if i == 0 {
				return "major", change
			} else if i == 1 {
				return "minor", change
			} else {
				return "patch", change
			}
		}
	}

	return "none", 0
}

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

func getCleanVersion(version string) (string, string) {
	re := regexp.MustCompile(`([^0-9]*)(\d+\.\d+\.\d+)(.*)`)
	reSubmatch := re.FindStringSubmatch(version)

	prefix := reSubmatch[1]
	cleanVersion := reSubmatch[2]

	return prefix, cleanVersion
}

func getRepositoryUrl(url string) string {
	// https://regex101.com/r/AEVsGf/2
	re := regexp.MustCompile(`(git[+@])?(.*)`)
	matches := re.FindStringSubmatch(url)

	if matches == nil {
		return ""
	}

	repoUrl := matches[2]
	repoUrl = strings.TrimSuffix(repoUrl, ".git")

	// https://regex101.com/r/iGLlG8/1
	re = regexp.MustCompile(`((ssh:\/\/git@)?(git:\/\/)?)(.*)`)
	matches = re.FindStringSubmatch(repoUrl)

	if matches == nil {
		return ""
	}

	repoUrl = matches[4]

	// Replace ":" with "/"
	repoUrl = strings.Replace(repoUrl, ".com:", ".com/", -1)

	// Append "https://" if missing
	if !strings.HasPrefix(repoUrl, "https://") {
		repoUrl = "https://" + repoUrl
	}

	return repoUrl
}

const concurrencyLimit int = 10

func readDependencies(
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
		versionPrefix, cleanCurrentVersion := getCleanVersion(currentVersion)

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
			resp, err := fetchNpmRegistry(dependency)
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
				repositoryUrl = getRepositoryUrl(repository["url"].(string))
			}

			// Get latest version from distTags
			var latestVersion string
			for key := range distTags {
				if key == "latest" {
					latestVersion = distTags[key].(string)
				}
			}

			// Get version update type (major, minor, patch, none)
			upgradeType, upgradeDirection := getVersionUpdateType(cleanCurrentVersion, latestVersion)

			// Save data
			if (upgradeDirection == 1) ||
				(cfg.AllowDowngrade && upgradeDirection == -1) {
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

func Init(cfg CmdFlags) {

	var isFilterFilled bool = cfg.Filter != ""

	fmt.Println()

	// Read json file
	jsonFile, err := os.ReadFile("package.json")
	if err != nil {
		fmt.Println(aurora.Red("No package.json found."), "Please run this command from the root of the project.")
		fmt.Println()
		return
	}

	// Parse json file
	var packageJsonMap PackageJSON
	err = json.Unmarshal(jsonFile, &packageJsonMap)
	if err != nil {
		fmt.Println(aurora.Red("Error reading package.json"), "invalid JSON or corrupt file. Error:")
		fmt.Println(err)
		fmt.Println()
		return
	}

	// Get dependencies
	dependencies := packageJsonMap.Dependencies
	devDependencies := packageJsonMap.DevDependencies

	if cfg.Dev {
		// Merge devDependencies with dependencies
		for key, value := range devDependencies {
			dependencies[key] = value
		}
	}

	versionComparison := map[string]VersionComparisonItem{}

	// Progress bar
	maxBar := len(dependencies)
	bar := initProgressBar(maxBar)

	// Process dependencies
	dependencyCount := len(dependencies)
	readDependencies(dependencies, versionComparison, false, bar, cfg)

	if cfg.Dev {
		dependencyCount += len(devDependencies)
		readDependencies(devDependencies, versionComparison, true, bar, cfg)
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

			freezeProgressCount := false

			response, err := promptUpdateDependency(
				updatePackageOptions,
				key,
				value.current,
				value.latest,
				value.versionType,
				updateProgressCount,
				totalCount,
			)

			if !freezeProgressCount {
				updateProgressCount++
			}

			if err != nil {
				if err == terminal.InterruptErr {
					log.Fatal("interrupted")
				}
			}

			if response == updatePackageOptions.skip {
				freezeProgressCount = true
				break
			}

			if response == updatePackageOptions.show_changes {
				freezeProgressCount = true

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
