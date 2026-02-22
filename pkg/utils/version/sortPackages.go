package version

import (
	"sort"
)

// PackageVersion represents a package with its name and version comparison item
type PackageVersion struct {
	Name string
	VersionComparisonItem
}

// SortPackagesByVersionType sorts packages by version type
func SortPackagesByVersionType(versionComparison map[string]VersionComparisonItem) []PackageVersion {
	packages := make([]PackageVersion, 0, len(versionComparison))

	// Convert map to slice
	for name, item := range versionComparison {
		packages = append(packages, PackageVersion{
			Name:                  name,
			VersionComparisonItem: item,
		})
	}

	// Sort by version type priority
	sort.Slice(packages, func(i, j int) bool {
		priorityI := getVersionTypePriority(packages[i].VersionType)
		priorityJ := getVersionTypePriority(packages[j].VersionType)
		return priorityI < priorityJ
	})

	return packages
}

// getVersionTypePriority returns priority for sorting: patch > minor > major
func getVersionTypePriority(versionType UpgradeType) int {
	switch versionType {
	case Patch:
		return 1
	case Minor:
		return 2
	case Major:
		return 3
	default:
		return 4
	}
}
