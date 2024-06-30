package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

type PackageJSON struct {
	Version string `json:"version"`
}

func GetPackageJsonVersion() string {
	// Open
	file, err := os.Open("package.json")
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	// Read
	bytes, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	// Deserialize JSON content
	var pkg PackageJSON
	err = json.Unmarshal(bytes, &pkg)
	if err != nil {
		log.Fatalf("Error deserializing el JSON: %v", err)
	}

	return pkg.Version
}
