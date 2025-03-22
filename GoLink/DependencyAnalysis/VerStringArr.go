package DependencyAnalysis

import (
	"sort"
	"strings"
)

// ByLengthAndLexicographical
type ByLengthAndLexicographical []string

func (a ByLengthAndLexicographical) Len() int {
	return len(a)
}

func (a ByLengthAndLexicographical) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByLengthAndLexicographical) Less(i, j int) bool {

	if len(a[i]) != len(a[j]) {
		return len(a[i]) < len(a[j])
	}

	return a[i] < a[j]
}

func extractPkgName(impName string) string {
	parts := strings.Split(impName, "/")
	if len(parts) < 3 {
		return impName
	}
	return strings.Join(parts[0:3], "/")
}

func extractSameVer(s1, s2 string) string {
	parts1 := strings.Split(s1, ", ")
	parts2 := strings.Split(s2, ", ")
	sameParts := intersect(parts1, parts2)
	return strings.Join(sameParts, ", ")
}

// intersect
func intersect(slice1, slice2 []string) []string {

	elementMap := make(map[string]bool)
	for _, item := range slice1 {
		elementMap[item] = true
	}
	var intersection []string
	for _, item := range slice2 {
		if elementMap[item] {
			intersection = append(intersection, item)
		}
	}

	sort.Sort(ByLengthAndLexicographical(intersection))

	return intersection
}
