package version

import "regexp"

func GetCleanVersion(version string) (string, string) {
	re := regexp.MustCompile(`([^0-9]*)(\d+\.\d+\.\d+)(.*)`)
	reSubmatch := re.FindStringSubmatch(version)

	prefix := reSubmatch[1]
	cleanVersion := reSubmatch[2]

	return prefix, cleanVersion
}
