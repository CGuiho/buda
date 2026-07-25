package qmd

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var versionPattern = regexp.MustCompile(`(?i)\bqmd\s+v?(\d+)\.(\d+)\.(\d+)\b`)

func ParseVersion(value string) (Version, error) {
	match := versionPattern.FindStringSubmatch(strings.TrimSpace(value))
	if match == nil {
		return Version{}, fmt.Errorf("parse qmd version from %q", strings.TrimSpace(value))
	}
	major, _ := strconv.Atoi(match[1])
	minor, _ := strconv.Atoi(match[2])
	patch, _ := strconv.Atoi(match[3])
	return Version{Raw: strings.TrimSpace(value), Major: major, Minor: minor, Patch: patch}, nil
}

func compareVersion(left, right Version) int {
	leftParts := [...]int{left.Major, left.Minor, left.Patch}
	rightParts := [...]int{right.Major, right.Minor, right.Patch}
	for index := range leftParts {
		if leftParts[index] < rightParts[index] {
			return -1
		}
		if leftParts[index] > rightParts[index] {
			return 1
		}
	}
	return 0
}

func parseConstraint(value string) (Version, error) {
	return ParseVersion("qmd " + value)
}
