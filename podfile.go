package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitrise-io/bitrise-init/scanners/ios"
	"github.com/bitrise-io/bitrise-init/utility"
	"github.com/bitrise-io/go-utils/pathutil"
)

type DepfileProvider interface {
	DepFilePath() (string, error)
	LockFilePath() (string, error)
}

type InputPodfileProvider struct {
	podfile string
}

func NewInputPodfileProvider(podfilePath string) DepfileProvider {
	return InputPodfileProvider{podfile: podfilePath}
}

func (p InputPodfileProvider) DepFilePath() (string, error) {
	depFilePath, err := pathutil.AbsPath(p.podfile)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(depFilePath); os.IsNotExist(err) {
		// https://stackoverflow.com/a/12518877
		return "", fmt.Errorf("%s is not exist", depFilePath)
	}

	return depFilePath, nil
}

func (p InputPodfileProvider) LockFilePath() (string, error) {
	podfile, err := p.DepFilePath()
	if err != nil {
		return "", err
	}

	dir := filepath.Dir(podfile)
	podfileLock := filepath.Join(dir, "Podfile.lock")
	if _, err := os.Stat(podfileLock); os.IsNotExist(err) {
		return "", fmt.Errorf("%s is not exist", podfileLock)
	}

	return podfileLock, nil
}

type PodDepFileProvider struct {
	dir string
}

func NewPodDepFileProvider(dir string) DepfileProvider {
	return PodDepFileProvider{dir: dir}
}

func (p PodDepFileProvider) DepFilePath() (string, error) {
	fileList, err := utility.ListPathInDirSortedByComponents(p.dir, false)
	if err != nil {
		return "", err
	}

	podfiles, err := utility.FilterPaths(fileList,
		ios.AllowPodfileBaseFilter,
		ios.ForbidCarthageDirComponentFilter,
		ios.ForbidPodsDirComponentFilter,
		ios.ForbidGitDirComponentFilter,
		ios.ForbidFramworkComponentWithExtensionFilter)
	if err != nil {
		return "", err
	}

	podfiles, err = utility.SortPathsByComponents(podfiles)
	if err != nil {
		return "", err
	}

	if len(podfiles) == 0 {
		return "", nil
	}

	return podfiles[0], nil
}

func (p PodDepFileProvider) LockFilePath() (string, error) {
	podfile, err := p.DepFilePath()
	if err != nil {
		return "", err
	}

	dir := filepath.Dir(podfile)
	podfileLock := filepath.Join(dir, "Podfile.lock")
	if _, err := os.Stat(podfileLock); os.IsNotExist(err) {
		return "", fmt.Errorf("%s is not exist", podfileLock)
	}

	return podfileLock, nil
}
