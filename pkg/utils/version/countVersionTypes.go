package version

type VersionComparisonItem struct {
	Current              string
	Latest               string
	VersionType          UpgradeType
	ShouldUpdate         bool
	Homepage             string
	RepositoryUrl        string
	VersionPrefix        string
	IsDev                bool
	HoursSinceLasRelease float64
}

func CountVersionTypes(
	versionComparison map[string]VersionComparisonItem,
) (
	majorCount int, minorCount int, patchCount int, totalCount int,
) {
	for _, value := range versionComparison {

		switch value.VersionType {
		case Major:
			majorCount++
		case Minor:
			minorCount++
		case Patch:
			patchCount++
		}

		totalCount++
	}

	return
}
