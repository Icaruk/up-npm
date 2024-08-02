package version

import (
	"regexp"
)

func GetCleanVersion(version string) (string, string) {

	if version == "" {
		return "", ""
	}

	re := regexp.MustCompile(`([^0-9]*)(\d?)\.?(\d?)\.?(\d?)(.*)`)
	reSubmatch := re.FindStringSubmatch(version) // [0] all, [1] = prefix, [2] = major, [3] = minor, [4] = patch

	prefix := reSubmatch[1]
	major := reSubmatch[2]
	minor := reSubmatch[3]
	patch := reSubmatch[4]

	if major == "" {
		major = "0"
	}
	if minor == "" {
		minor = "0"
	}
	if patch == "" {
		patch = "0"
	}

	cleanVersion := major + "." + minor + "." + patch

	return prefix, cleanVersion
}
