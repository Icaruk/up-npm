package updater

import (
	"fmt"
	"os"

	"github.com/icaruk/updatenpm/pkg/updater"
	"github.com/spf13/cobra"
)

var version = "0.0.1"
var dev bool

var rootCmd = &cobra.Command{
	Use:     "updatenpm",
	Version: version,
	Short:   "Updates npm depeendencies",
	Long:    `updatenpm is a easy way to keep your npm depeendencies up to date.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		devFlag, err := cmd.Flags().GetBool("dev")
		if err != nil {
			return err
		}
		updater.Init(devFlag)
		return nil
	},
}

func init() {
	rootCmd.Flags().BoolVarP(&dev, "dev", "d", false, "Update dev dependencies")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
