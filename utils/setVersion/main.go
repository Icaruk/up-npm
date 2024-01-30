package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {

	var newVersion string
	flag.StringVar(&newVersion, "version", "", "New version to set")
	flag.Parse()

	if newVersion == "" {
		fmt.Println("No --version provided")
		return
	}

	fmt.Println("Setting version:", newVersion)

	filePath := "./cmd/updater/root.go"

	// Open file for reading and writing
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Create scanner to read file line by line
	scanner := bufio.NewScanner(file)

	// Create a slice to store the lines of the file
	var lines []string

	// Iterate over each line of the file
	for scanner.Scan() {
		line := scanner.Text()

		// Search line with "__VERSION__"
		if strings.Contains(line, "const __VERSION__ string =") {
			// Replace whole line
			line = fmt.Sprintf("const __VERSION__ string = \"%s\"", newVersion) // const __VERSION__ string = "1.2.3"
		}

		// Add line to slice
		lines = append(lines, line)
	}

	// Truncate(0) will empty the file
	if err := file.Truncate(0); err != nil {
		fmt.Println("Error truncating file:", err)
		return
	}

	// Write back the modified lines to the file
	if _, err := file.Seek(0, 0); err != nil {
		fmt.Println("Error al buscar al inicio del archivo:", err)
		return
	}

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}
	writer.Flush()

	fmt.Println("Version updated sucessfully from ", filePath)
}
