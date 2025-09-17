package cli

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	versionpkg "github.com/icaruk/up-npm/pkg/utils/version"
	"github.com/logrusorgru/aurora/v4"
)

type SelectUpdateOptions struct {
	update       string
	skip         string
	show_changes string
	finish       string
}

var SelectUpdateAvailableOptions = SelectUpdateOptions{
	update:       "Update",
	skip:         "Skip",
	show_changes: "Show changes",
	finish:       "Finish",
}

func PromptUpdateDependency(
	dependencyName string,
	versionComparisonItem versionpkg.VersionComparisonItem,
	currentCount int,
	maxCount int,
) string {

	var selected string

	lockedVersionWarning := ""
	tooRecentReleaseWarning := ""

	if versionComparisonItem.VersionPrefix == "" {
		lockedVersionWarning = aurora.Sprintf("\n%s", aurora.Faint("version is locked"))
	}

	if versionComparisonItem.HoursSinceLasRelease > 0 && versionComparisonItem.HoursSinceLasRelease < 24 {
		tooRecentReleaseWarning = aurora.Sprintf(
			"\n%s%s %g %s",
			aurora.Bold(aurora.Red("WARNING: ")),
			aurora.Red("has been released"),
			aurora.Yellow(versionComparisonItem.HoursSinceLasRelease),
			aurora.Red("hours ago!"),
		)
	}

	selectForm := huh.NewSelect[string]().
		Title(
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("7")). // white
				PaddingTop(1).
				Render(
					fmt.Sprintf(
						"[%d/%d] Update \"%s\" from %s to %s?%s%s",
						currentCount,
						maxCount,
						dependencyName,
						versionComparisonItem.Current,
						versionpkg.ColorizeVersion(versionComparisonItem.Latest, versionComparisonItem.VersionType),
						lockedVersionWarning,
						tooRecentReleaseWarning,
					),
				),
		).
		Options(
			huh.NewOption(SelectUpdateAvailableOptions.update, SelectUpdateAvailableOptions.update),
			huh.NewOption(SelectUpdateAvailableOptions.skip, SelectUpdateAvailableOptions.skip),
			huh.NewOption(SelectUpdateAvailableOptions.show_changes, SelectUpdateAvailableOptions.show_changes),
			huh.NewOption(SelectUpdateAvailableOptions.finish, SelectUpdateAvailableOptions.finish),
		).
		Value(&selected).
		WithTheme(huh.ThemeBase16())

	selectForm.Run()

	return selected
}
