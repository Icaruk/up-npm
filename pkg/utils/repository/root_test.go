package repository

import (
	"testing"
)

func TestGetRepositoryUrl(t *testing.T) {
	testCases := []struct {
		url         string
		expectedUrl string
	}{
		{url: "https://github.com/ghuser/package-name", expectedUrl: "https://github.com/ghuser/package-name"},
		{url: "https://github.com/ghuser/package-name.git", expectedUrl: "https://github.com/ghuser/package-name"},
		{url: "https://github.com/ghuser/package-name.git#main", expectedUrl: "https://github.com/ghuser/package-name"},

		{url: "git+https://github.com/ghuser/package-name", expectedUrl: "https://github.com/ghuser/package-name"},
		{url: "git+https://github.com/ghuser/package-name.git", expectedUrl: "https://github.com/ghuser/package-name"},
		{url: "git+https://github.com/ghuser/package-name.git#main", expectedUrl: "https://github.com/ghuser/package-name"},

		{url: "git@github.com:ghuser/package-name", expectedUrl: "https://github.com/ghuser/package-name"},
		{url: "git@github.com:ghuser/package-name.git", expectedUrl: "https://github.com/ghuser/package-name"},
		{url: "git@github.com:ghuser/package-name.git#main", expectedUrl: "https://github.com/ghuser/package-name"},

		{url: "ssh://git@github.com/mongodb/node-mongodb-native", expectedUrl: "https://github.com/mongodb/node-mongodb-native"},
		{url: "ssh://git@github.com/mongodb/node-mongodb-native.git", expectedUrl: "https://github.com/mongodb/node-mongodb-native"},
		{url: "ssh://git@github.com/mongodb/node-mongodb-native.git#main", expectedUrl: "https://github.com/mongodb/node-mongodb-native"},
	}

	for _, tc := range testCases {
		t.Run(tc.url, func(t *testing.T) {
			returnedUrl := GetRepositoryUrl(tc.url)
			if returnedUrl != tc.expectedUrl {
				t.Errorf("expectedUrl %v but got %v for %v", tc.expectedUrl, returnedUrl, tc.url)
			}
		})
	}
}

func TestGetRepositoryUrlMetadata(t *testing.T) {
	testCases := []struct {
		url                    string
		expectedUsername       string
		expectedRepositoryName string
	}{
		{url: "https://github.com/ghuser/package-name", expectedUsername: "ghuser", expectedRepositoryName: "package-name"},
	}

	for _, tc := range testCases {
		t.Run(tc.url, func(t *testing.T) {
			urlMetadata := GetRepositoryUrlMetadata(tc.url)

			if urlMetadata.username != tc.expectedUsername || urlMetadata.repositoryName != tc.expectedRepositoryName {
				t.Errorf(
					"expectedUsername %v and expectedRepositoryName %v but got %v and %v for %v",
					tc.expectedUsername,
					tc.expectedRepositoryName,
					urlMetadata.username,
					urlMetadata.repositoryName,
					tc.url,
				)
			}
		})
	}
}
