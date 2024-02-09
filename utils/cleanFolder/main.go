package main

import (
	"flag"
	"fmt"
	"os"
	"path"
)

func main() {

	var targetPath string
	flag.StringVar(&targetPath, "path", "", "Path to the folder to clean")
	flag.Parse()

	if targetPath == "" {
		fmt.Println("No --path provided")
		return
	}

	fmt.Println("Cleaning:", targetPath)

	// Delete all files in the folder recursively
	dir, _ := os.ReadDir(targetPath)
	for _, d := range dir {
		// fmt.Println(d)
		os.RemoveAll(path.Join([]string{targetPath, d.Name()}...))
	}

	fmt.Println("Done")
}
