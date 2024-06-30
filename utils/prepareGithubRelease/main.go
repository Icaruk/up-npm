package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	method := "POST"

	version := GetPackageJsonVersion()
	fmt.Printf("Version: %s\n", version)

	payload := map[string]interface{}{
		"tag_name":               version,
		"target_commitish":       "main",
		"name":                   version,
		"body":                   "Test automated release",
		"draft":                  true,
		"prerelease":             false,
		"generate_release_notes": false,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Println(err)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, "https://api.github.com/repos/Icaruk/up-npm/releases", bytes.NewBuffer(jsonPayload))

	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("Authorization", "Bearer <YOUR-TOKEN>")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Get "upload_url" property from JSON body
	var decodedBody map[string]interface{}
	err = json.Unmarshal(body, &decodedBody)
	if err != nil {
		fmt.Println(err)
		return
	}

	uploadURL := decodedBody["upload_url"].(string)
	fmt.Printf("uploadURL: %s\n", uploadURL)

	// ----------------------------
	// Upload assets
	// ----------------------------

	// Perform POST request
	req, err = http.NewRequest("POST", uploadURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("Authorization", "Bearer <YOUR-TOKEN>")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	res, err = client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err = io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(body))

}
