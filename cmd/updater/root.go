package updater

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/icaruk/updatenpm/pkg/updater"
	"github.com/spf13/cobra"
)

var dev bool

var rootCmd = &cobra.Command{
	Use:   "up-npm",
	Short: "Updates npm depeendencies",
	Long:  `up-npm is a easy way to keep your npm depeendencies up to date.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		devFlag, err := cmd.Flags().GetBool("dev")
		if err != nil {
			return err
		}

		filterFlag, err := cmd.Flags().GetString("filter")
		if err != nil {
			return err
		}

		updater.Init(devFlag, filterFlag)
		return nil
	},
}

func init() {
	rootCmd.Flags().BoolVarP(&dev, "dev", "d", false, "Include dev dependencies")
	rootCmd.Flags().StringP("filter", "f", "", "Filter dependencies by package name")

	binaryPath, err := os.Executable()
	if err != nil {
		fmt.Println(err)
	}

	localVersionPath := filepath.Join(filepath.Dir(binaryPath), "../version")

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
