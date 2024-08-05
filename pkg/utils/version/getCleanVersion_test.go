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
		{
			version:         "^15.0.0-canary.102",
			expectedPrefix:  "^",
			expectedVersion: "15.0.0",
		},
		{
			version:         "^1",
			expectedPrefix:  "^",
			expectedVersion: "1.0.0",
		},
		{
			version:         "^1.2",
			expectedPrefix:  "^",
			expectedVersion: "1.2.0",
		},
		{
			version:         "12.34.56",
			expectedPrefix:  "",
			expectedVersion: "12.34.56",
		},
		{
			version:         "123.3456",
			expectedPrefix:  "",
			expectedVersion: "123.3456.0",
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
