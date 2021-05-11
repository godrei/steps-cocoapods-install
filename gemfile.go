package main

import (
	"errors"
	"os"
	"path/filepath"
)

type GemFileProvider struct {
	dir string
}

func NewGemFileProvider(dir string) DepfileProvider {
	return GemFileProvider{dir: dir}
}

func (p GemFileProvider) DepFilePath() (string, error) {
	// gems.rb, gems.locked
	// https://github.com/rubygems/bundler/issues/694
	for _, gemFileName := range []string{"gems.rb", "Gemfile"} {
		pth := filepath.Join(p.dir, gemFileName)
		if _, err := os.Stat(pth); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", err
		}
		return pth, nil
	}
	return "", errors.New("Gemfile not found")
}

func (p GemFileProvider) LockFilePath() (string, error) {
	// gems.rb, gems.locked
	// https://github.com/rubygems/bundler/issues/694
	for _, gemFileLockName := range []string{"Gemfile.lock", "gems.locked"} {
		pth := filepath.Join(p.dir, gemFileLockName)
		if _, err := os.Stat(pth); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", err
		}
		return pth, nil
	}
	return "", errors.New("Gemfile lock not found")
}
