package version

import (
	"regexp"
	"strconv"
)

func GetVersionComponents(semver string) (int, int, int) {
	re := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)
	matches := re.FindStringSubmatch(semver)
	vCurrentMajor, _ := strconv.Atoi(matches[1])
	vCurrentMinor, _ := strconv.Atoi(matches[2])
	vCurrentPatch, _ := strconv.Atoi(matches[3])

	return vCurrentMajor, vCurrentMinor, vCurrentPatch
}
