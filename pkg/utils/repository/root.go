package repository

import (
	"regexp"
	"strings"
)

func GetRepositoryUrl(url string) string {
	// https://regex101.com/r/AEVsGf/2
	re := regexp.MustCompile(`(git[+@])?(.*)`)
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
