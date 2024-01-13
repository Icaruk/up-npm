package version

import (
	"strconv"
	"strings"
)

// UpgradeType enum
type UpgradeType string

const (
	Major UpgradeType = "major"
	Minor UpgradeType = "minor"
	Patch UpgradeType = "patch"
	NoneT UpgradeType = "none"
)

// UpgradeDirection enum
type UpgradeDirection string

const (
	Upgrade   UpgradeDirection = "upgrade"
	Downgrade UpgradeDirection = "downgrade"
	None      UpgradeDirection = "none"
)

func GetVersionUpdateType(currentVersion, latestVersion string) (upgradeType UpgradeType, upgradeDirection UpgradeDirection) {

	if latestVersion == currentVersion {
		return "none", "none"
	}

	currentArr := strings.Split(currentVersion, ".") // 1.0.0
	latestArr := strings.Split(latestVersion, ".")   // 0.9.9

	versionChange := []int{0, 0, 0}

	for i := 0; i < len(currentArr); i++ {
		vCur, _ := strconv.Atoi(currentArr[i])
		vLat, _ := strconv.Atoi(latestArr[i])

		// 0 = major
		// 1 = minor
		// 2 = patch

		var changeDirection int

		if vLat > vCur {
			changeDirection = 1
		} else if vLat < vCur {
			changeDirection = -1
		}

		versionChange[i] = changeDirection

	}

	for i, change := range versionChange {
		if change == -1 {
			if i == 0 {
				return "major", "downgrade"
			} else if i == 1 {
				return "minor", "downgrade"
			} else {
				return "patch", "downgrade"
			}
		}
		if change == 1 {
			if i == 0 {
				return "major", "upgrade"
			} else if i == 1 {
				return "minor", "upgrade"
			} else {
				return "patch", "upgrade"
			}
		}
	}

	return "none", "none"
}
