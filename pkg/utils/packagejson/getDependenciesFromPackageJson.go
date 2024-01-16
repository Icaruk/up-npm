package packagejson

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/logrusorgru/aurora/v4"
)

type PackageJSON struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Homepage        string            `json:"homepage"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

func GetDependenciesFromPackageJson(packageJsonFilename string, preventDevDependencies bool) (dependencies map[string]string, jsonFile []byte, err error) {
	// Read json file
	jsonFile, err = os.ReadFile(packageJsonFilename)
	if err != nil {
		errStr := aurora.Sprintf(
			aurora.Red("No package.json found."),
			"Please run this command from the root of the project.",
		)
		return nil, nil, errors.New(errStr)
	}

	// Parse json file
	var packageJsonMap PackageJSON

	err = json.Unmarshal(jsonFile, &packageJsonMap)
	if err != nil {
		errStr := aurora.Sprintf(
			aurora.Red("Error reading package.json"),
			"Invalid JSON or corrupt file. Error:",
			err,
		)
		return nil, nil, errors.New(errStr)
	}

	// Initialize empty
	dependencies = make(map[string]string)
	devDependencies := make(map[string]string)

	// Get dependencies
	if packageJsonMap.Dependencies != nil {
		dependencies = packageJsonMap.Dependencies
	}
	if packageJsonMap.DevDependencies != nil {
		devDependencies = packageJsonMap.DevDependencies
	}

	if dependencies == nil && devDependencies == nil {
		errStr := aurora.Sprintf(
			aurora.Red("No dependencies found on package.json"),
		)

		return nil, nil, errors.New(errStr)
	}

	if !preventDevDependencies {
		// Merge devDependencies with dependencies
		for key, value := range devDependencies {
			dependencies[key] = value
		}
	}

	return dependencies, jsonFile, nil
}
