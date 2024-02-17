package repository

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

func GetRepositoryUrl(url string) string {

	// Remove #main #something at the end of the url
	re := regexp.MustCompile(`#.*`)
	url = re.ReplaceAllString(url, "${1}")

	// https://regex101.com/r/AEVsGf/2
	re = regexp.MustCompile(`(git[+@])?(.*)`)
	matches := re.FindStringSubmatch(url)

	if matches == nil {
		return ""
	}

	repoUrl := matches[2]
	repoUrl = strings.TrimSuffix(repoUrl, ".git")

	// https://regex101.com/r/iGLlG8/1
	re = regexp.MustCompile(`((ssh:\/\/git@)?(git:\/\/)?)(.*)`)
	matches = re.FindStringSubmatch(repoUrl)

	if matches == nil {
		return ""
	}

	repoUrl = matches[4]

	// Replace ":" with "/"
	repoUrl = strings.Replace(repoUrl, ".com:", ".com/", -1)

	// Append "https://" if missing
	if !strings.HasPrefix(repoUrl, "https://") {
		repoUrl = "https://" + repoUrl
	}

	return repoUrl
}

type URLMetadata struct {
	Username       string
	RepositoryName string
}

// GetRepositoryUrlMetadata retrieves the username and repository name from the given URL.
//
// url: string - the URL to extract username and repository name from
// URLMetadata - the struct containing Username and RepositoryName
func GetRepositoryUrlMetadata(url string) URLMetadata {
	// URL is like https://github.com/username/repo-name

	// Split with "/"
	fragments := strings.Split(url, "/")

	// Get last 2 fragments
	username := fragments[len(fragments)-2]
	repositoryName := fragments[len(fragments)-1]

	return URLMetadata{
		Username:       username,
		RepositoryName: repositoryName,
	}

}

/*
Get latest release from a given repository
*/
func FetchRepositoryLatestRelease(user string, repository string) (map[string]interface{}, error) {

	// Build URL like https://api.github.com/repos/<user>/<repository>/releases
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", user, repository)

	res, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	// Check for successful status code (200 OK)
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", res.StatusCode)
	}

	// Attempt to decode body as JSON into a map[string]interface{}
	var decodedBody map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&decodedBody)
	if err != nil {
		// Handle JSON decoding error gracefully
		if strings.Contains(err.Error(), "unexpected EOF") {
			return nil, fmt.Errorf("empty JSON response from GitHub API")
		} else {
			return nil, fmt.Errorf("error decoding JSON body: %w", err)
		}
	}

	return decodedBody, nil

}

/*
Get CHANGELOG.md file from a given repository
*/
func FetchRepositoryChangelogFile(user string, repository string) (map[string]interface{}, error) {

	// Build URL like https://api.github.com/repos/<user>/<repository>/contents/CHANGELOG.md
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/CHANGELOG.md", user, repository)

	res, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	// Check for successful status code (200 OK)
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", res.StatusCode)
	}

	// Attempt to decode body as JSON into a map[string]interface{}
	var decodedBody map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&decodedBody)
	if err != nil {
		// Handle JSON decoding error gracefully
		if strings.Contains(err.Error(), "unexpected EOF") {
			return nil, fmt.Errorf("empty JSON response from GitHub API")
		} else {
			return nil, fmt.Errorf("error decoding JSON body: %w", err)
		}
	}

	return decodedBody, nil

}
