package cli

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	versionpkg "github.com/icaruk/up-npm/pkg/utils/version"
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
	versionFrom string,
	versionTo string,
	versionType versionpkg.UpgradeType,
) string {

	var selected string

	selectForm := huh.NewSelect[string]().
		Title(
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("7")). // white
				PaddingTop(1).
				Render(
					fmt.Sprintf(
						"Update \"%s\" from %s to %s?",
						dependencyName,
						versionFrom,
						versionpkg.ColorizeVersion(versionTo, versionType),
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
