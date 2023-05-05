package updater

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/icaruk/updatenpm/pkg/updater"
	"github.com/spf13/cobra"
)

var Cfg = updater.CmdFlags{
	Dev:            false,
	AllowDowngrade: false,
	Filter:         "",
}

var rootCmd = &cobra.Command{
	Use:   "up-npm",
	Short: "Updates npm dependencies",
	Long:  `up-npm is a easy way to keep your npm dependencies up to date.`,
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

		Cfg = updater.CmdFlags{
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

var whereCmd = &cobra.Command{
	Use:   "where",
	Short: "Prints where up-npm is installed",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get current location
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		fmt.Println(wd)

		return nil
	},
}

func init() {
	rootCmd.Flags().BoolVarP(&Cfg.Dev, "dev", "d", false, "Include dev dependencies")
	rootCmd.Flags().StringVarP(&Cfg.Filter, "filter", "f", "", "Filter dependencies by package name")
	rootCmd.Flags().BoolVar(&Cfg.AllowDowngrade, "allow-downgrade", false, "Allows downgrading a if latest version is older than current")

	rootCmd.AddCommand(whereCmd)

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
