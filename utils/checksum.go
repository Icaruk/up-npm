package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	hashes := make(chan string)

	// Create the output file as a buffered writer
	outputFileName := "hashes.sha256"
	outputPath := "dist/" + outputFileName

	outputFile, err := os.Create(outputPath)
	if err != nil {
		fmt.Printf("Error creating file %s: %s\n", outputFileName, err)
		return
	}
	defer outputFile.Close()
	writer := bufio.NewWriter(outputFile)

	// Walk through the "dist" directory and calculate the SHA256 checksum for each file
	err = filepath.Walk("dist", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {

			// Skip the output file
			if filepath.Base(path) == outputFileName {
				return nil
			}

			wg.Add(1)

			go func() {
				defer wg.Done()

				// Open the file
				file, err := os.Open(path)
				if err != nil {
					fmt.Printf("Error opening file %s: %s\n", path, err)
					return
				}
				defer file.Close()

				// Create the SHA256 hash
				hash := sha256.New()

				// Read the file and calculate the hash
				if _, err := io.Copy(hash, file); err != nil {
					fmt.Printf("Error reading file %s: %s\n", path, err)
					return
				}

				// Get the hash result in hexadecimal format
				hashString := hex.EncodeToString(hash.Sum(nil))

				// Send the hash to the channel
				hashes <- fmt.Sprintf("%s *%s", hashString, filepath.Base(path))

				fmt.Printf("Checksum for file %s has been calculated\n", path)
			}()
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	// Close the channel when all the hashes have been received
	go func() {
		wg.Wait()
		close(hashes)
	}()

	// Write the hashes to the output file in the order they are received
	for hash := range hashes {
		_, err := writer.WriteString(fmt.Sprintf("%s\n", hash))
		if err != nil {
			fmt.Printf("Error writing to file %s: %s\n", outputFileName, err)
			return
		}
	}

	// Flush the buffer to ensure all data is written to the file
	writer.Flush()

	fmt.Printf("\nDone. All checksums have been calculated\n")
}
