package updater

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/icaruk/up-npm/pkg/utils/version"
)

func TestPrintReport(t *testing.T) {
	packages := []version.PackageVersion{
		{Name: "pkg-b", VersionComparisonItem: version.VersionComparisonItem{Current: "1.0.0", Latest: "2.0.0", VersionType: version.Major}},
		{Name: "pkg-a", VersionComparisonItem: version.VersionComparisonItem{Current: "1.0.0", Latest: "2.0.0", VersionType: version.Major}},
		{Name: "pkg-c", VersionComparisonItem: version.VersionComparisonItem{Current: "1.0.0", Latest: "1.0.1", VersionType: version.Patch}},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	printReport(packages)
	w.Close()
	os.Stdout = oldStdout

	out, _ := io.ReadAll(r)
	s := string(out)

	if !strings.Contains(s, "Major updates: 2") || !strings.Contains(s, "Patch updates: 1") {
		t.Errorf("missing section titles:\n%s", s)
	}
	if strings.Contains(s, "Minor updates") {
		t.Errorf("empty section should be omitted:\n%s", s)
	}
	if strings.Index(s, "Major updates") > strings.Index(s, "Patch updates") {
		t.Errorf("Major should render before Patch:\n%s", s)
	}
}

func TestGetVersionUpdateType(t *testing.T) {
	testCases := []struct {
		name              string
		current           string
		latest            string
		expectedType      version.UpgradeType
		expectedDirection version.UpgradeDirection
	}{
		{name: "Test Case 1", current: "1.0.0", latest: "1.0.1", expectedType: "patch", expectedDirection: "upgrade"},
		{name: "Test Case 2", current: "1.0.0", latest: "1.1.0", expectedType: "minor", expectedDirection: "upgrade"},
		{name: "Test Case 3", current: "1.0.0", latest: "2.0.0", expectedType: "major", expectedDirection: "upgrade"},
		{name: "Test Case 4", current: "2.0.0", latest: "2.0.0", expectedType: "none", expectedDirection: "none"},
		{name: "Test Case 5", current: "1.9.9", latest: "2.1.2", expectedType: "major", expectedDirection: "upgrade"},
		{name: "Test Case 6", current: "3.5.6", latest: "3.6.9", expectedType: "minor", expectedDirection: "upgrade"},
		{name: "Test Case 7", current: "20.10.10", latest: "20.10.11", expectedType: "patch", expectedDirection: "upgrade"},
		{name: "Test Case 8", current: "1.1.1", latest: "0.1.1", expectedType: "major", expectedDirection: "downgrade"},
		{name: "Test Case 9", current: "1.1.1", latest: "1.0.1", expectedType: "minor", expectedDirection: "downgrade"},
		{name: "Test Case 10", current: "1.1.1", latest: "1.1.0", expectedType: "patch", expectedDirection: "downgrade"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			upgradeType, upgadeDirection := version.GetVersionUpdateType(tc.current, tc.latest)
			if upgradeType != tc.expectedType {
				t.Errorf("expectedType %v but got %v for %v and %v", tc.expectedType, upgradeType, tc.current, tc.latest)
			}
			if upgadeDirection != tc.expectedDirection {
				t.Errorf("expectedDirection %v but got %v for %v and %v", tc.expectedDirection, upgadeDirection, tc.current, tc.latest)
			}
		})
	}
}

func TestGetVersionComponents(t *testing.T) {
	testCases := []struct {
		name          string
		semver        string
		expectedMajor int
		expectedMinor int
		expectedPatch int
	}{
		{name: "Test Case 1", semver: "1.0.0", expectedMajor: 1, expectedMinor: 0, expectedPatch: 0},
		{name: "Test Case 2", semver: "2.1.5", expectedMajor: 2, expectedMinor: 1, expectedPatch: 5},
		{name: "Test Case 3", semver: "0.0.1", expectedMajor: 0, expectedMinor: 0, expectedPatch: 1},
		{name: "Test Case 4", semver: "30.12.17", expectedMajor: 30, expectedMinor: 12, expectedPatch: 17},
		{name: "Test Case 5", semver: "0.10.0", expectedMajor: 0, expectedMinor: 10, expectedPatch: 0},
		{name: "Test Case 6", semver: "1.2.3-beta.1", expectedMajor: 1, expectedMinor: 2, expectedPatch: 3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			major, minor, patch := version.GetVersionComponents(tc.semver)
			if major != tc.expectedMajor || minor != tc.expectedMinor || patch != tc.expectedPatch {
				t.Errorf("Expected %v.%v.%v but got %v.%v.%v for %v", tc.expectedMajor, tc.expectedMinor, tc.expectedPatch, major, minor, patch, tc.semver)
			}
		})
	}
}
