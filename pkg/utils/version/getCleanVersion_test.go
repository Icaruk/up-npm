package version

import (
	"testing"
)

func TestGetCleanVersion(t *testing.T) {
	testCases := []struct {
		version         string
		expectedPrefix  string
		expectedVersion string
	}{
		{
			version:         "",
			expectedPrefix:  "",
			expectedVersion: "",
		},
		{
			version:         "^1.2.3",
			expectedPrefix:  "^",
			expectedVersion: "1.2.3",
		},
		{
			version:         "~1.2.0",
			expectedPrefix:  "~",
			expectedVersion: "1.2.0",
		},
		{
			version:         "1.2.3",
			expectedPrefix:  "",
			expectedVersion: "1.2.3",
		},
		{
			version:         "^1.2.3-alpha.1",
			expectedPrefix:  "^",
			expectedVersion: "1.2.3",
		},
	}

	for _, tc := range testCases {
		t.Run(string(tc.version), func(t *testing.T) {
			prefix, version := GetCleanVersion(tc.version)
			if prefix != tc.expectedPrefix || version != tc.expectedVersion {
				t.Errorf("version %v: expected %v and %v but %v and %v", version, tc.expectedPrefix, tc.expectedVersion, prefix, version)
			}
		})
	}
}
