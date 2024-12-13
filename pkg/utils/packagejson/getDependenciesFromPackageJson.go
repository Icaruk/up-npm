package packagejson

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/logrusorgru/aurora/v4"
)

type PackageJSON struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Homepage        string            `json:"homepage"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

func GetDependenciesFromPackageJson(
	packageJsonFilename string,
	preventDevDependencies bool,
) (
	dependencies map[string]string,
	devDependencies map[string]string,
	jsonFile []byte,
	err error,
) {

	// Read json file
	jsonFile, err = os.ReadFile(packageJsonFilename)
	if err != nil {

		var b strings.Builder

		fmt.Fprintf(
			&b,
			aurora.Red("File \"%s\" not found.").String(),
			packageJsonFilename,
		)

		errStr := b.String()

		return nil, nil, nil, errors.New(errStr)
	}

	// Parse json file
	var packageJsonMap PackageJSON

	err = json.Unmarshal(jsonFile, &packageJsonMap)
	if err != nil {

		var b strings.Builder

		fmt.Fprintf(
			&b,
			aurora.Red("Error reading file \"%s\".\n").String(),
			packageJsonFilename,
		)
		fmt.Fprintf(&b, "Invalid JSON or corrupt file.")
		fmt.Fprintf(&b, "\nError: %s", err)

		errStr := b.String()

		return nil, nil, nil, errors.New(errStr)
	}

	// Initialize empty
	dependencies = make(map[string]string)
	devDependencies = make(map[string]string)

	// Get dependencies
	if packageJsonMap.Dependencies != nil {
		dependencies = packageJsonMap.Dependencies
	}
	if packageJsonMap.DevDependencies != nil {
		devDependencies = packageJsonMap.DevDependencies
	}

	if len(dependencies) == 0 && len(devDependencies) == 0 {
		errStr := aurora.Sprintf(
			aurora.Red("No dependencies found on file \"%s\"."),
			packageJsonFilename,
		)

		return nil, nil, nil, errors.New(errStr)
	}

	return dependencies, devDependencies, jsonFile, nil
}
