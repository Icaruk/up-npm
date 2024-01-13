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
	NoDev:          false,
	AllowDowngrade: false,
	Filter:         "",
}

type Flag struct {
	Long  string
	Short string
}

var AllowedFlags = map[string]Flag{
	"noDev": {
		Long: "no-dev",
	},
	"filter": {
		Long:  "filter",
		Short: "f",
	},
	"allowDowngrade": {
		Long: "allow-downgrade",
	},
}

var rootCmd = &cobra.Command{
	Use:   "up-npm",
	Short: "Updates npm depeendencies",
	Long:  `up-npm is a easy way to keep your npm depeendencies up to date.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		noDevFlag, err := cmd.Flags().GetBool(AllowedFlags["noDev"].Long)
		if err != nil {
			return err
		}

		filterFlag, err := cmd.Flags().GetString(AllowedFlags["filter"].Long)
		if err != nil {
			return err
		}

		allowDowngradeFlag, err := cmd.Flags().GetBool(AllowedFlags["allowDowngrade"].Long)
		if err != nil {
			return err
		}

		Cfg = npm.CmdFlags{
			NoDev:          noDevFlag,
			Filter:         filterFlag,
			AllowDowngrade: allowDowngradeFlag,
		}

		// Cfg.NoDev = noDevFlag
		// Cfg.AllowDowngrade = noDevFlag

		updater.Init(Cfg)

		return nil
	},
}

func init() {
	rootCmd.Flags().BoolVar(
		&Cfg.NoDev,
		AllowedFlags["noDev"].Long,
		false,
		"Exclude dev dependencies",
	)
	rootCmd.Flags().StringVarP(&Cfg.Filter,
		AllowedFlags["filter"].Long,
		AllowedFlags["filter"].Short,
		"",
		"Filter dependencies by package name",
	)
	rootCmd.Flags().BoolVar(&Cfg.AllowDowngrade,
		AllowedFlags["allowDowngrade"].Long,
		false,
		"Allows downgrading a if latest version is older than current",
	)

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
