package updater

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/icaruk/up-npm/pkg/utils/cli"
	npm "github.com/icaruk/up-npm/pkg/utils/npm"
	"github.com/icaruk/up-npm/pkg/utils/npmrc"
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

func Init(cfg npm.CmdFlags, binVersion string) {

	var isFilterFilled bool = cfg.Filter != ""

	// Check new version
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

	// Check .npmrc
	npmrcFiles, _ := npmrc.GetNpmrcTokens()
	token, npmrcTokenLevel := npmrc.GetRelevantNpmrcToken(npmrcFiles)

	if token != "" {

		fmt.Println(
			aurora.Green(".npmrc").Hyperlink("https://docs.npmjs.com/cli/v10/configuring-npm/npmrc"),
			aurora.Green("has been detected"),
			aurora.Faint(fmt.Sprintf("(%s)", npmrcTokenLevel)),
		)

		fmt.Println()

	}

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

	lockedDependencyCount = npm.FetchDependencies(dependencies, versionComparison, false, token, bar, cfg)

	// Process devDependencies
	if !cfg.NoDev {
		lockedDevDependencyCount = npm.FetchDependencies(devDependencies, versionComparison, true, token, bar, cfg)
	}

	// Count total dependencies and filtered dependencies
	filteredDependencyCount := totalDependencyCount
	if isFilterFilled {
		filteredDependencyCount = len(versionComparison)
	}

	// Count version types
	majorCount, minorCount, patchCount, totalCount := versionpkg.CountVersionTypes(versionComparison)
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

	currentUpdateCount := 1
	maxUpdateCount := len(versionComparison)

	for key, value := range versionComparison {

		exit := false

		for {

			if cfg.UpdatePatches {

				if value.VersionType == versionpkg.Patch {
					// get a copy of the entry
					if entry, ok := versionComparison[key]; ok {
						entry.ShouldUpdate = true      // then modify the copy
						versionComparison[key] = entry // then reassign map entry
					}

					colorizedVersion := versionpkg.ColorizeVersion(value.Latest, value.VersionType)

					fmt.Println(
						aurora.Sprintf(
							"%s \"%s\" from %s to %s",
							aurora.Green("Auto updated"),
							key,
							value.Current,
							colorizedVersion,
						),
					)

					break
				}
			}

			response := cli.PromptUpdateDependency(
				key,
				value.Current,
				value.Latest,
				value.VersionType,
				currentUpdateCount,
				maxUpdateCount,
			)

			if response == updatePackageOptions.skip {
				// Skipped dependencyName in green color
				fmt.Println(
					aurora.Sprintf(
						aurora.Faint("Skipped \"%s\""),
						key,
					),
				)

				currentUpdateCount++

				break
			}

			if response == updatePackageOptions.show_changes {

				if value.RepositoryUrl == "" {
					fmt.Println(aurora.Red("Repository URL does not exist"))
					continue
				}

				// Get user and repository from repository URL
				urlMetadata := repositorypkg.GetRepositoryUrlMetadata(value.RepositoryUrl)

				// Fetch repository from github
				_, err := repositorypkg.FetchRepositoryLatestRelease(urlMetadata.Username, urlMetadata.RepositoryName)
				missingLatestRelease := err != nil

				var changelogMdUrl string

				if missingLatestRelease {
					fmt.Println(aurora.Faint("Latest release from github does not exist"))

					// Fetch CHANGELOG.md
					response, err := repositorypkg.FetchRepositoryChangelogFile(urlMetadata.Username, urlMetadata.RepositoryName)

					if err == nil {
						changelogMdUrl = response["html_url"].(string)
					} else {
						fmt.Println(aurora.Faint("CHANGELOG.md does not exist"))
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
					continue
				}

				fmt.Println("Opening...")
				fmt.Println()
				cli.Openbrowser(url)

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

				currentUpdateCount++

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

	packageManager := packagejson.GetPackageManager()
	installationCommand := packagejson.GetInstallationCommand(packageManager)

	installPromptMessage := fmt.Sprintf("Run '%s' to install dependencies?", installationCommand)
	response, err = cli.PromptYesNo(installPromptMessage)

	if err != nil {
		if err == terminal.InterruptErr {
			fmt.Println("")
		}
	}

	if response == cli.YesNoPromptOptions.Yes {
		fmt.Printf("Running '%s'...", installationCommand)

		// Split command and args
		commandAndArgs := strings.Split(installationCommand, " ")
		command := commandAndArgs[0]
		args := commandAndArgs[1:]

		// Execute command
		cmd := exec.Command(command, args...)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Println(err)
		}

		return
	}

}
