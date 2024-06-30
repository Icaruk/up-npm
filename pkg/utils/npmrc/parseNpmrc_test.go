package npmrc

import (
	"testing"
)

func TestParseNpmrc(t *testing.T) {
	testCases := []struct {
		testName     string
		npmrcContent string
		expected     string
	}{
		{
			testName:     "empty token",
			npmrcContent: "",
			expected:     "",
		},
		{
			testName:     "correct token",
			npmrcContent: "//registry.npmjs.org/:_authToken=npm_1234",
			expected:     "npm_1234",
		},
		{
			testName:     "incomplete authToken",
			npmrcContent: "//registry.npmjs.org/:_authToken=",
			expected:     "",
		},
		{
			testName:     "wrong format authToken",
			npmrcContent: "registry.npmjs.org/authToken=npm_1234",
			expected:     "",
		},
		{
			testName:     "commented authToken with ;",
			npmrcContent: "; registry.npmjs.org/authToken=npm_1234",
			expected:     "",
		},
		{
			testName:     "commented authToken with #",
			npmrcContent: "# registry.npmjs.org/authToken=npm_1234",
			expected:     "",
		},
	}

	for _, tc := range testCases {
		t.Run(string(tc.testName), func(t *testing.T) {
			token, _ := ParseNpmrc(tc.npmrcContent)
			if token != tc.expected {
				t.Errorf("case %v but got %v", tc.testName, token)
			}
		})
	}
}
