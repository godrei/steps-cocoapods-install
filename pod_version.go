package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bitrise-io/go-utils/fileutil"
)

func cocoapodsVersionFromPodfileLockContent(content string) string {
	exp := regexp.MustCompile("COCOAPODS: (.+)")
	match := exp.FindStringSubmatch(content)
	if len(match) == 2 {
		return match[1]
	}
	return ""
}

func cocoapodsVersionFromPodfileLock(podfileLockPth string) (string, error) {
	content, err := fileutil.ReadStringFromFile(podfileLockPth)
	if err != nil {
		return "", err
	}
	return cocoapodsVersionFromPodfileLockContent(content), nil
}

// VersionSpec ...
type VersionSpec struct {
	Operator string
	Version  string
}

func splitOperatorAndVersion(input string) (VersionSpec, error) {
	splittedString := strings.Split(input, " ")
	cnt := len(splittedString)

	if cnt == 1 {
		out := VersionSpec{"", splittedString[0]}
		return out, nil
	}

	if cnt != 2 {
		err := fmt.Errorf("Invalid version range: %s", input)
		return VersionSpec{}, err
	}

	out := VersionSpec{splittedString[0], splittedString[1]}
	return out, nil
}

func isIncludedInGemfileLockVersionRanges(input string, gemfileLockVersion string) (bool, error) {
	var splittedVersions = strings.Split(gemfileLockVersion, ", ")

	for _, each := range splittedVersions {
		versionSpec, err := splitOperatorAndVersion(each)
		if err != nil {
			return false, err
		}

		switch versionSpec.Operator {
		case "":
			if input != versionSpec.Version {
				return false, nil
			}

			continue
		case "~>":
			if input != versionSpec.Version {
				return false, nil
			}

			continue
		case ">=":
			versions := strings.Split(versionSpec.Version, ".")
			inputVersions := strings.Split(input, ".")

			for i, version := range versions {
				v1, err := strconv.Atoi(version)
				if err != nil {
					return false, err
				}

				v2, err := strconv.Atoi(inputVersions[i])
				if err != nil {
					return false, err
				}

				if i != len(versions)-1 && v1 == v2 {
					continue
				}
				if v2 >= v1 {
					break
				} else {
					return false, nil
				}
			}

			continue
		case "<":
			versions := strings.Split(versionSpec.Version, ".")
			inputVersions := strings.Split(input, ".")

			for i, version := range versions {
				v1, err := strconv.Atoi(version)
				if err != nil {
					return false, err
				}

				v2, err := strconv.Atoi(inputVersions[i])
				if err != nil {
					return false, err
				}

				if i != len(versions)-1 && v1 == v2 {
					continue
				}
				if v2 < v1 {
					break
				} else {
					return false, nil
				}
			}

			continue
		default:
			err := fmt.Errorf("Unknown version operator: %s", each)
			return false, err
		}
	}

	return true, nil
}
