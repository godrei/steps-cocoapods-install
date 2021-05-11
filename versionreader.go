package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	"github.com/bitrise-io/go-steputils/command/gems"
)

type DepVersionComparator interface {
	Included(version, versionDef string) (bool, error)
}

type GemDepVersionComparator struct {
}

func NewGemDepVersionComparator() DepVersionComparator {
	return GemDepVersionComparator{}
}

func (c GemDepVersionComparator) Included(version, versionDef string) (bool, error) {
	var splittedVersions = strings.Split(versionDef, ", ")

	for _, each := range splittedVersions {
		versionSpec, err := splitOperatorAndVersion(each)
		if err != nil {
			return false, err
		}

		switch versionSpec.Operator {
		case "":
			if version != versionSpec.Version {
				return false, nil
			}

			continue
		case "~>":
			if version != versionSpec.Version {
				return false, nil
			}

			continue
		case ">=":
			versions := strings.Split(versionSpec.Version, ".")
			inputVersions := strings.Split(version, ".")

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
			inputVersions := strings.Split(version, ".")

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

type DepVersionReader interface {
	ReadVersion(r io.Reader) (string, error)
}

type PodfileLockPodVersionReader struct {
}

func NewPodfileLockPodVersionReader() DepVersionReader {
	return PodfileLockPodVersionReader{}
}

func (versionReader PodfileLockPodVersionReader) ReadVersion(r io.Reader) (string, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}

	exp := regexp.MustCompile("COCOAPODS: (.+)")
	match := exp.FindStringSubmatch(string(b))
	if len(match) == 2 {
		return match[1], nil
	}
	return "", nil
}

type GemFileLockPodVersionReader struct {
}

func NewGemFileLockVersionReader() DepVersionReader {
	return GemFileLockPodVersionReader{}
}

func (versionReader GemFileLockPodVersionReader) ReadVersion(r io.Reader) (string, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}

	pod, err := gems.ParseVersionFromBundle("cocoapods", string(b))
	if err != nil {
		failf("Failed to check if gem lockfile contains cocoapods, error: %s", err)
	}

	if pod.Found {
		return pod.Version, nil
	}
	return "", nil
}

type GemfileLockBundleVersionReader struct {
}

func NewGemfileLockBundleVersionReader() GemfileLockBundleVersionReader {
	return GemfileLockBundleVersionReader{}
}

func (versionReader GemfileLockBundleVersionReader) ReadVersion(r io.Reader) (string, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}

	bundler, err := gems.ParseBundlerVersion(string(b))
	if err != nil {
		failf("Failed to parse bundler version form cocoapods, error: %s", err)
	}

	if bundler.Found {
		return bundler.Version, nil
	}
	return "", nil
}
