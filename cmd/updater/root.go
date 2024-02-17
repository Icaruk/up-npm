package updater

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/icaruk/up-npm/pkg/updater"
	"github.com/icaruk/up-npm/pkg/utils/npm"
	"github.com/spf13/cobra"
)

const __VERSION__ string = "4.2.0"

var Cfg = npm.CmdFlags{
	NoDev:          false,
	AllowDowngrade: false,
	Filter:         "",
	File:           "",
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
	"file": {
		Long: "file",
	},
}

var rootCmd = &cobra.Command{
	Use:   "up-npm",
	Short: "Updates npm dependencies",
	Long:  `up-npm is a easy way to keep your npm dependencies up to date.`,
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

		file, err := cmd.Flags().GetString(AllowedFlags["file"].Long)
		if err != nil {
			return err
		}

		Cfg = npm.CmdFlags{
			NoDev:          noDevFlag,
			Filter:         filterFlag,
			AllowDowngrade: allowDowngradeFlag,
			File:           file,
		}

		updater.Init(Cfg)

		return nil
	},
}

var whereCmd = &cobra.Command{
	Use:   "where",
	Short: "Prints where up-npm is installed",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get bin location
		binPath, err := os.Executable()
		if err != nil {
			return err
		}

		fmt.Println(filepath.Dir(binPath))

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
	rootCmd.Flags().StringVarP(
		&Cfg.Filter,
		AllowedFlags["filter"].Long,
		AllowedFlags["filter"].Short,
		"",
		"Filter dependencies by package name",
	)
	rootCmd.Flags().BoolVar(
		&Cfg.AllowDowngrade,
		AllowedFlags["allowDowngrade"].Long,
		false,
		"Allows downgrading a if latest version is older than current",
	)
	rootCmd.Flags().StringVar(
		&Cfg.File,
		AllowedFlags["file"].Long,
		"package.json",
		"File dependencies by package name",
	)

	rootCmd.AddCommand(whereCmd)

	rootCmd.Version = string(__VERSION__)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
