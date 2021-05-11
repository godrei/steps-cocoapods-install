package main

import (
	"github.com/bitrise-io/go-steputils/command/rubycommand"
	"github.com/bitrise-io/go-utils/log"
)

type InstallCommandGenerator interface {
	IsInstalled(version string) (bool, error)
	InstallCommand(version string) [][]string
	RequiredVersion() (string, error)
}

type RubyInstallCommandGenerator struct {
	dir string
}

func NewRubyInstallCommandGenerator(workDir string) InstallCommandGenerator {
	return RubyInstallCommandGenerator{dir: workDir}
}

func (g RubyInstallCommandGenerator) InstallCommand(version string) [][]string {
	return [][]string{[]string{"rbenv", "install", version}}
}

func (g RubyInstallCommandGenerator) IsInstalled(_ string) (bool, error) {
	rubyInstalled, _, err := rubycommand.IsSpecifiedRbenvRubyInstalled(g.dir)
	if err != nil {
		log.Errorf("Failed to check if selected ruby is installed, error: %s", err)
	}
	return rubyInstalled, nil
}

func (g RubyInstallCommandGenerator) RequiredVersion() (string, error) {
	_, rubyVersion, err := rubycommand.IsSpecifiedRbenvRubyInstalled(g.dir)
	if err != nil {
		log.Errorf("Failed to check if selected ruby is installed, error: %s", err)
	}
	return rubyVersion, nil
}

type RubyPodInstallCommandGenerator struct {
}

func NewPodInstallCommandGenerator() InstallCommandGenerator {
	return RubyPodInstallCommandGenerator{}
}

func (g RubyPodInstallCommandGenerator) InstallCommand(version string) [][]string {
	cmds, err := rubycommand.GemInstall("cocoapods", version, false)
	if err != nil {

	}

	var c [][]string
	for _, cmd := range cmds {
		c = append(c, cmd.GetCmd().Args)
	}
	return c
}

func (g RubyPodInstallCommandGenerator) IsInstalled(version string) (bool, error) {
	return rubycommand.IsGemInstalled("cocoapods", version)
}

func (g RubyPodInstallCommandGenerator) RequiredVersion() (string, error) {
	return "", nil
}
