package updater

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/icaruk/up-npm/pkg/updater"
	"github.com/icaruk/up-npm/pkg/utils/npm"
	"github.com/spf13/cobra"
)

var Cfg = npm.CmdFlags{
	Dev:            false,
	AllowDowngrade: false,
	Filter:         "",
}

var rootCmd = &cobra.Command{
	Use:   "up-npm",
	Short: "Updates npm depeendencies",
	Long:  `up-npm is a easy way to keep your npm depeendencies up to date.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		devFlag, err := cmd.Flags().GetBool("dev")
		if err != nil {
			return err
		}

		allowDowngradeFlag, err := cmd.Flags().GetBool("dev")
		if err != nil {
			return err
		}

		filterFlag, err := cmd.Flags().GetString("filter")
		if err != nil {
			return err
		}

		Cfg = npm.CmdFlags{
			Dev:            devFlag,
			AllowDowngrade: allowDowngradeFlag,
			Filter:         filterFlag,
		}

		Cfg.Dev = devFlag
		Cfg.AllowDowngrade = devFlag

		updater.Init(Cfg)

		return nil
	},
}

func init() {
	rootCmd.Flags().BoolVar(&Cfg.Dev, "no-dev", false, "Include dev dependencies")
	rootCmd.Flags().StringVarP(&Cfg.Filter, "filter", "f", "", "Filter dependencies by package name")
	rootCmd.Flags().BoolVar(&Cfg.AllowDowngrade, "allow-downgrade", false, "Allows downgrading a if latest version is older than current")

	binaryPath, err := os.Executable()
	if err != nil {
		fmt.Println(err)
	}

	localVersionPath := filepath.Join(filepath.Dir(binaryPath), "../.version")

	localVersion := "0.0.0"
	localVersionByte, err := os.ReadFile(localVersionPath)
	if err == nil {
		localVersion = string(localVersionByte)
	}

	rootCmd.Version = string(localVersion)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
