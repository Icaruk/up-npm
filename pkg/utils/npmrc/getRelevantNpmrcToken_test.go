package npmrc

import (
	"testing"
)

func TestGetRelevantNpmrcToken(t *testing.T) {
	testCases := []struct {
		should     string
		npmrcToken NpmrcTokens
		expected   NpmrcTokenLevel
	}{
		{
			should: "return empty string",
			npmrcToken: NpmrcTokens{
				Exists:  false,
				Project: "",
				User:    "",
				Global:  "",
				Builtin: "",
			},
			expected: "",
		},
		{
			should: "return Project",
			npmrcToken: NpmrcTokens{
				Exists:  true,
				Project: "npm_token",
				User:    "",
				Global:  "",
				Builtin: "",
			},
			expected: Project,
		},
		{
			should: "return User",
			npmrcToken: NpmrcTokens{
				Exists:  true,
				Project: "",
				User:    "npm_token",
				Global:  "",
				Builtin: "",
			},
			expected: User,
		},
		{
			should: "return Global",
			npmrcToken: NpmrcTokens{
				Exists:  true,
				Project: "",
				User:    "",
				Global:  "npm_token",
				Builtin: "",
			},
			expected: Global,
		},
		{
			should: "return Builtin",
			npmrcToken: NpmrcTokens{
				Exists:  true,
				Project: "",
				User:    "",
				Global:  "",
				Builtin: "npm_token",
			},
			expected: Builtin,
		},
	}

	for _, tc := range testCases {
		t.Run(string(tc.should), func(t *testing.T) {
			_, tokenLevel := GetRelevantNpmrcToken(tc.npmrcToken)
			if tokenLevel != tc.expected {
				t.Errorf("should %v but got %v", tc.should, tokenLevel)
			}
		})
	}
}
