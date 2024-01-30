package version

import (
	"github.com/logrusorgru/aurora/v4"
)

func ColorizeVersion(version string, versionType UpgradeType) string {
	vLatestMajor, vLatestMinor, vLatestPatch := GetVersionComponents(version)

	var colorizedVersion string

	if versionType == Major {
		colorizedVersion = aurora.Sprintf(
			aurora.Red("%d.%d.%d"),
			aurora.Red(vLatestMajor),
			aurora.Red(vLatestMinor),
			aurora.Red(vLatestPatch),
		)
	} else if versionType == Minor {
		colorizedVersion = aurora.Sprintf(aurora.White("%d.%d.%d"), aurora.White(vLatestMajor), aurora.Yellow(vLatestMinor), aurora.Yellow(vLatestPatch))
	} else if versionType == Patch {
		colorizedVersion = aurora.Sprintf(aurora.White("%d.%d.%d"), aurora.White(vLatestMajor), aurora.White(vLatestMinor), aurora.Green(vLatestPatch))
	}

	return colorizedVersion
}
