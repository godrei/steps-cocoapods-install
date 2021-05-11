package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitrise-io/go-steputils/cache"
	"github.com/bitrise-io/go-steputils/command/gems"
	"github.com/bitrise-io/go-steputils/command/rubycommand"
	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
)

func failf(format string, v ...interface{}) {
	log.Errorf(format, v...)
	os.Exit(1)
}

func main() {
	configs, err := createConfigsModelFromEnvs()
	if err != nil {
		failf(err.Error())
	}

	stepconf.Print(configs)

	//
	fmt.Println()
	log.Infof("Searching for Podfile")

	var podfileProvider DepfileProvider
	if configs.PodfilePath != "" {
		podfileProvider = NewInputPodfileProvider(configs.PodfilePath)
	} else {
		podfileProvider = NewPodDepFileProvider(configs.SourceRootPath)
	}
	podfilePath, err := podfileProvider.DepFilePath()
	if err != nil {
		failf("Failed to find Podfile: %s", err)
	}
	if podfilePath == "" {
		failf("No Podfile found")
	}

	log.Donef("Using Podfile: %s", podfilePath)

	podfileLockPth, err := podfileProvider.LockFilePath()
	if err != nil {
		log.Warnf("No Podfile.lock found: %s", err)
		log.Warnf("Make sure it's committed into your repository!")
	} else {
		log.Donef("Using Podfile.lock: %s", podfilePath)
	}

	var gemFileProvider DepfileProvider
	gemFileProvider = NewGemFileProvider(filepath.Dir(podfilePath))
	gemFilePath, err := gemFileProvider.DepFilePath()
	if err != nil {
		log.Warnf("No Gemfile found: %s", err)
	} else {
		log.Donef("Using Gemfile: %s", gemFilePath)
	}

	gemFileLockPath, err := gemFileProvider.LockFilePath()
	if err != nil {
		log.Warnf("No Gemfile.lock found: %s", err)
	} else {
		log.Donef("Using Gemfile.lock: %s", gemFileLockPath)
	}

	podVersion := ""
	if podfileLockPth != "" {
		var podfileLockPodVersionReader DepVersionReader
		podfileLockPodVersionReader = NewPodfileLockPodVersionReader()
		podfileLock, err := os.Open(podfileLockPth)
		if err != nil {

		}
		podVersion, err = podfileLockPodVersionReader.ReadVersion(podfileLock)
		if err != nil {

		}
	}

	useBundler := false
	if gemFileLockPath != "" {
		var gemFileLockPodVersionReader DepVersionReader
		gemFileLockPodVersionReader = NewGemFileLockVersionReader()
		gemFileLock, err := os.Open(gemFileLockPath)
		if err != nil {

		}
		version, err := gemFileLockPodVersionReader.ReadVersion(gemFileLock)
		if err != nil {

		}

		if version != "" {
			var depVersionComparator DepVersionComparator
			depVersionComparator = NewGemDepVersionComparator()
			isIncludedVersionRange, err := depVersionComparator.Included(podVersion, version)
			if err != nil {
				failf("Failed to compare version range in gem lockfile, error: %s", err)
			}
			if !isIncludedVersionRange {
				log.Warnf("Cocoapods version required in Podfile.lock (%s) does not match Gemfile.lock (%s). Will install Cocoapods using bundler.", useCocoapodsVersionFromPodfileLock, useCocoapodsVersionFromGemfileLock)
			}

			podVersion = version
			useBundler = true
		}
	}

	///

	podfileDir := filepath.Dir(podfilePath)

	//
	// Install required cocoapods version
	fmt.Println()
	log.Infof("Determining required cocoapods version")

	useBundler := false
	useCocoapodsVersionFromPodfileLock := ""
	useCocoapodsVersionFromGemfileLock := ""

	log.Printf("Searching for Podfile.lock")

	// Check Podfile.lock for CocoaPods version
	podfileLockPth, err := podfileProvider.LockFilePath()
	if err == nil {
		// Podfile.lock exist search for version
		log.Printf("Found Podfile.lock: %s", podfileLockPth)

		version, err := cocoapodsVersionFromPodfileLock(podfileLockPth)
		if err != nil {
			failf("Failed to determine CocoaPods version, error: %s", err)
		}

		if version != "" {
			useCocoapodsVersionFromPodfileLock = version
			log.Donef("Required CocoaPods version (from Podfile.lock): %s", useCocoapodsVersionFromPodfileLock)
		} else {
			log.Warnf("No CocoaPods version found in Podfile.lock! (%s)", podfileLockPth)
		}
	} else {
		log.Warnf("No Podfile.lock found at: %s", podfileLockPth)
		log.Warnf("Make sure it's committed into your repository!")
	}

	var pod gems.Version
	var bundler gems.Version

	log.Printf("Searching for gem lockfile with cocoapods gem")

	// Check gem lockfile for CocoaPods version
	gemfileLockPth, err := gems.GemFileLockPth(podfileDir)
	if err != nil && err != gems.ErrGemLockNotFound {
		failf("Failed to check gem lockfile at: %s, error: %s", podfileDir, err)
	}

	if gemfileLockPth != "" {
		// CocoaPods exist search for version in gem lockfile
		log.Printf("Found gem lockfile: %s", gemfileLockPth)

		content, err := fileutil.ReadStringFromFile(gemfileLockPth)
		if err != nil {
			failf("failed to read file (%s) contents, error: %s", gemfileLockPth, err)
		}

		pod, err = gems.ParseVersionFromBundle("cocoapods", content)
		if err != nil {
			failf("Failed to check if gem lockfile contains cocoapods, error: %s", err)
		}

		bundler, err = gems.ParseBundlerVersion(content)
		if err != nil {
			failf("Failed to parse bundler version form cocoapods, error: %s", err)
		}

		if pod.Found {
			useCocoapodsVersionFromGemfileLock = pod.Version
			log.Donef("Required CocoaPods version (from gem lockfile): %s", useCocoapodsVersionFromGemfileLock)

			isIncludedVersionRange, err := isIncludedInGemfileLockVersionRanges(useCocoapodsVersionFromPodfileLock, useCocoapodsVersionFromGemfileLock)
			if err != nil {
				failf("Failed to compare version range in gem lockfile, error: %s", err)
			}

			if !isIncludedVersionRange {
				log.Warnf("Cocoapods version required in Podfile.lock (%s) does not match Gemfile.lock (%s). Will install Cocoapods using bundler.", useCocoapodsVersionFromPodfileLock, useCocoapodsVersionFromGemfileLock)
			}
			useBundler = true
		}
	} else {
		log.Printf("No gem lockfile with cocoapods gem found at: %s", gemfileLockPth)
		log.Donef("Using system installed CocoaPods version")
	}

	// Check ruby version
	// Run this logic only in CI environment when the ruby was installed via rbenv for the virtual machine
	if os.Getenv("CI") == "true" && rubycommand.RubyInstallType() == rubycommand.RbenvRuby {
		fmt.Println()
		log.Infof("Check selected Ruby is installed")

		rubyInstalled, rversion, err := rubycommand.IsSpecifiedRbenvRubyInstalled(configs.SourceRootPath)
		if err != nil {
			log.Errorf("Failed to check if selected ruby is installed, error: %s", err)
		}

		if !rubyInstalled {
			log.Errorf("Ruby %s is not installed", rversion)
			fmt.Println()

			cmd := command.New("rbenv", "install", rversion).SetStdout(os.Stdout).SetStderr(os.Stderr)
			log.Donef("$ %s", cmd.PrintableCommandArgs())
			if err := cmd.Run(); err != nil {
				log.Errorf("Failed to install Ruby version %s, error: %s", rversion, err)
			}
		} else {
			log.Donef("Ruby %s is installed", rversion)
		}

	}

	// Install cocoapods
	fmt.Println()
	log.Infof("Installing cocoapods")

	podCmdSlice := []string{"pod"}

	if useBundler {
		fmt.Println()
		log.Infof("Installing bundler")

		// install bundler with `gem install bundler [-v version]`
		// in some configurations, the command "bunder _1.2.3_" can return 'Command not found', installing bundler solves this
		installBundlerCommand := gems.InstallBundlerCommand(bundler)
		installBundlerCommand.SetStdout(os.Stdout).SetStderr(os.Stderr)
		installBundlerCommand.SetDir(podfileDir)

		log.Donef("$ %s", installBundlerCommand.PrintableCommandArgs())
		fmt.Println()

		if err := installBundlerCommand.Run(); err != nil {
			failf("command failed, error: %s", err)
		}

		// install gem lockfile gems with `bundle [_version_] install ...`
		fmt.Println()
		log.Infof("Installing cocoapods with bundler")

		cmd, err := gems.BundleInstallCommand(bundler)
		if err != nil {
			failf("failed to create bundle command model, error: %s", err)
		}
		cmd.SetStdout(os.Stdout).SetStderr(os.Stderr)
		cmd.SetDir(podfileDir)

		log.Donef("$ %s", cmd.PrintableCommandArgs())
		fmt.Println()

		if err := cmd.Run(); err != nil {
			failf("Command failed, error: %s", err)
		}

		if useBundler {
			podCmdSlice = append(gems.BundleExecPrefix(bundler), podCmdSlice...)
		}
	} else if useCocoapodsVersionFromPodfileLock != "" {
		log.Printf("Checking cocoapods %s gem", useCocoapodsVersionFromPodfileLock)

		installed, err := rubycommand.IsGemInstalled("cocoapods", useCocoapodsVersionFromPodfileLock)
		if err != nil {
			failf("Failed to check if cocoapods %s installed, error: %s", useCocoapodsVersionFromPodfileLock, err)
		}

		if !installed {
			log.Printf("Installing")

			cmds, err := rubycommand.GemInstall("cocoapods", useCocoapodsVersionFromPodfileLock, false)
			if err != nil {
				failf("Failed to create command model, error: %s", err)
			}

			for _, cmd := range cmds {
				log.Donef("$ %s", cmd.PrintableCommandArgs())

				cmd.SetDir(podfileDir)

				if err := cmd.Run(); err != nil {
					failf("Command failed, error: %s", err)
				}
			}
		} else {
			log.Printf("Installed")
		}

		podCmdSlice = append(podCmdSlice, fmt.Sprintf("_%s_", useCocoapodsVersionFromPodfileLock))
	} else {
		log.Printf("Using system installed cocoapods")
	}

	fmt.Println()
	log.Infof("cocoapods version:")

	// pod can be in the PATH as an rbenv shim and pod --version will return "rbenv: pod: command not found"
	cmd, err := rubycommand.NewFromSlice(append(podCmdSlice, "--version"))
	if err != nil {
		failf("Failed to create command model, error: %s", err)
	}

	cmd.SetStdout(os.Stdout).SetStderr(os.Stderr)
	cmd.SetDir(podfileDir)

	log.Donef("$ %s", cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		failf("command failed, error: %s", err)
	}

	// Run pod install
	fmt.Println()
	log.Infof("Installing Pods")

	podInstallCmdSlice := append(podCmdSlice, "install", "--no-repo-update")
	if configs.Verbose == "true" {
		podInstallCmdSlice = append(podInstallCmdSlice, "--verbose")
	}

	cmd, err = rubycommand.NewFromSlice(podInstallCmdSlice)
	if err != nil {
		failf("Failed to create command model, error: %s", err)
	}

	cmd.SetStdout(os.Stdout).SetStderr(os.Stderr)
	cmd.SetDir(podfileDir)

	log.Donef("$ %s", cmd.PrintableCommandArgs())
	if err := cmd.Run(); err != nil {
		log.Warnf("Command failed, error: %s, retrying without --no-repo-update ...", err)

		// Repo update
		cmd, err = rubycommand.NewFromSlice(append(podCmdSlice, "repo", "update"))
		if err != nil {
			failf("Failed to create command model, error: %s", err)
		}

		cmd.SetStdout(os.Stdout).SetStderr(os.Stderr)
		cmd.SetDir(podfileDir)

		log.Donef("$ %s", cmd.PrintableCommandArgs())
		if err := cmd.Run(); err != nil {
			failf("Command failed, error: %s", err)
		}

		// Pod install
		podInstallCmdSlice := append(podCmdSlice, "install")
		if configs.Verbose == "true" {
			podInstallCmdSlice = append(podInstallCmdSlice, "--verbose")
		}

		cmd, err = rubycommand.NewFromSlice(podInstallCmdSlice)
		if err != nil {
			failf("Failed to create command model, error: %s", err)
		}

		cmd.SetStdout(os.Stdout).SetStderr(os.Stderr)
		cmd.SetDir(podfileDir)

		log.Donef("$ %s", cmd.PrintableCommandArgs())
		if err := cmd.Run(); err != nil {
			failf("Command failed, error: %s", err)
		}
	}

	// Collecting caches
	if configs.IsCacheDisabled != "true" && podfileLockPth != "" {
		fmt.Println()
		log.Infof("Collecting Pod cache paths...")

		podsCache := cache.New()
		podsCache.IncludePath(fmt.Sprintf("%s -> %s", filepath.Join(podfileDir, "Pods"), podfileLockPth))

		if err := podsCache.Commit(); err != nil {
			log.Warnf("Cache collection skipped: failed to commit cache paths.")
		}
	}

	log.Donef("Success!")
}
