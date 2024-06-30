package npmrc

import (
	"fmt"
	"os"
	"path/filepath"
)

const npmrcFilename = ".npmrc"

type NpmrcTokens struct {
	Exists  bool
	Project string
	User    string
	Global  string
	Builtin string
}

func GetNpmrcTokens() (NpmrcTokens, error) {

	/*
		The four relevant files are:

		· per-project config file (/path/to/my/project/.npmrc)
		· per-user config file (~/.npmrc)
		· [TODO] global config file ($PREFIX/etc/npmrc)
		· [TODO] npm builtin config file (/path/to/npm/npmrc)
	*/

	npmrcTokens := NpmrcTokens{
		Exists:  false,
		Project: "",
		User:    "",
		Global:  "",
		Builtin: "",
	}

	// per-project config file (/path/to/my/project/.npmrc)
	if _, err := os.Stat(npmrcFilename); err == nil {

		fileContent, err := os.ReadFile(npmrcFilename)
		if err != nil {
			fmt.Println(err)
		}

		if token, err := ParseNpmrc(string(fileContent)); err == nil {
			npmrcTokens.Exists = true
			npmrcTokens.Project = token
		}
	}

	// per-user config file (~/.npmrc)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
	}

	userNpmrcFilename := filepath.Join(homeDir, npmrcFilename)
	if _, err := os.Stat(userNpmrcFilename); err == nil {

		fileContent, err := os.ReadFile(userNpmrcFilename)
		if err != nil {
			fmt.Println(err)
		}

		if token, err := ParseNpmrc(string(fileContent)); err == nil {
			npmrcTokens.Exists = true
			npmrcTokens.User = token
		}
	}

	return npmrcTokens, nil

}
