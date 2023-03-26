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
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/jedib0t/go-pretty/v6/table"
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

func printSummary(totalCount int, patchCount int, minorCount int, majorCount int) {

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
	totalCount int, majorCount int, minorCount int, patchCount int,
) {
	for _, value := range versionComparison {
		totalCount++

		if value.versionType == "major" {
			majorCount++
		} else if value.versionType == "minor" {
			minorCount++
		} else if value.versionType == "patch" {
			patchCount++
		}
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

func promptUpdateDependency(options updatePackageOptions, dependency, currentVersion, latestVersion, versionType string) (string, error) {
	response := ""
	prompt := &survey.Select{
		Message: fmt.Sprintf(
			"Update \"%s\" from %s to %s?",
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

func promptWriteJson(options writeJsonOptions) string {
	response := ""
	prompt := &survey.Select{
		Message: "Update package.json?",
		Options: []string{
			options.yes,
			options.yes_backup,
			options.no,
		},
	}
	survey.AskOne(prompt, &response)

	return response
}

func getVersionType(currentVersion, latestVersion string) string {
	vCurrentMajor, vCurrentMinor, vCurrentPatch := getVersionComponents(currentVersion)
	vLatestMajor, vLatestMinor, vLatestPatch := getVersionComponents(latestVersion)

	var versionType string

	if vLatestMajor > vCurrentMajor {
		versionType = "major"
	} else if vLatestMinor > vCurrentMinor {
		versionType = "minor"
	} else if vLatestPatch > vCurrentPatch {
		versionType = "patch"
	}

	return versionType
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

func readDependencies(dependencyList map[string]string, targetMap map[string]VersionComparisonItem, isDev bool, bar *progressbar.ProgressBar) {

	if !isDev {
		isDev = false
	}

	for dependency, currentVersion := range dependencyList {

		// Get version and prefix
		versionPrefix, cleanCurrentVersion := getCleanVersion(currentVersion)

		if cleanCurrentVersion == "" {
			fmt.Println(dependency, " has unsupported/invalid version, skipping...")
			continue
		}

		// Perform get request to npm registry
		resp, err := fetchNpmRegistry(dependency)
		if err != nil {
			fmt.Println("Failed to fetch", dependency, " from npm registry, skipping...")
			continue
		}
		defer resp.Body.Close()

		// Get response data
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		distTags := result["dist-tags"].(map[string]interface{})
		homepage := result["homepage"].(string)

		// Get latest version from distTags
		var latestVersion string
		for key := range distTags {
			if key == "latest" {
				latestVersion = distTags[key].(string)
			}
		}

		// Get version type (major, minor, patch)
		versionType := getVersionType(cleanCurrentVersion, latestVersion)

		// Save data
		if versionType != "" {
			targetMap[dependency] = VersionComparisonItem{
				current:       cleanCurrentVersion,
				latest:        latestVersion,
				versionType:   versionType,
				shouldUpdate:  false,
				homepage:      homepage,
				versionPrefix: versionPrefix,
				isDev:         isDev,
			}
		}

		if bar != nil {
			bar.Add(1)
		}

	}

}

func Init(updateDev bool) {

	fmt.Println()

	// Read json file
	jsonFile, err := os.ReadFile("package.json")
	if err != nil {
		fmt.Println("No package.json found. Please run this command from the root of the project.")
	}

	// Parse json file
	var packageJsonMap PackageJSON
	err = json.Unmarshal(jsonFile, &packageJsonMap)
	if err != nil {
		fmt.Println(err)
	}

	// Get dependencies
	dependencies := packageJsonMap.Dependencies
	devDependencies := packageJsonMap.DevDependencies

	if updateDev {
		// Merge devDependencies with dependencies
		for key, value := range devDependencies {
			dependencies[key] = value
		}
	}

	versionComparison := map[string]VersionComparisonItem{}

	// Progress bar
	maxBar := len(dependencies)
	bar := initProgressBar(maxBar)

	readDependencies(dependencies, versionComparison, false, bar)
	if updateDev {
		readDependencies(devDependencies, versionComparison, true, bar)
	}

	// Table
	fmt.Println("")
	fmt.Println("")
	printUpdatablePackagesTable(versionComparison)
	fmt.Println("")

	// Count version types
	totalCount, majorCount, minorCount, patchCount := countVersionTypes(versionComparison)
	printSummary(totalCount, majorCount, minorCount, patchCount)
	fmt.Println()

	// Prompt user to update each dependency
	updatePackageOptions := updatePackageOptions{
		update:       "Update",
		skip:         "Skip",
		show_changes: "Show changes",
		finish:       "Finish",
	}

	for key, value := range versionComparison {

		exit := false

		for {

			response, err := promptUpdateDependency(updatePackageOptions, key, value.current, value.latest, value.versionType)

			if err != nil {
				if err == terminal.InterruptErr {
					log.Fatal("interrupted")
				}
			}

			if response == updatePackageOptions.skip {
				fmt.Println("Skipping update for \"", key, "\"")
				break
			}

			if response == updatePackageOptions.show_changes {
				// Open browser url
				homepage := value.homepage

				if homepage == "" {
					fmt.Println("No homepage found")
				} else {
					fmt.Println("Opening homepage...")
					fmt.Println()
					Openbrowser(homepage)
				}

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

	// Prompt to write package.json
	writeJsonOptions := writeJsonOptions{
		yes:        "Yes",
		yes_backup: "Yes, but backup before update",
		no:         "No",
	}

	response := promptWriteJson(writeJsonOptions)

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
			if updateDev {
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
	fmt.Println(aurora.Green("Updated package.json"))
	fmt.Println("Run 'npm install' to install dependencies")

}
