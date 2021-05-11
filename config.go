package main

import (
	"fmt"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-utils/pathutil"
)

// ConfigsModel ...
type ConfigsModel struct {
	SourceRootPath  string `env:"source_root_path,dir"`
	PodfilePath     string `env:"podfile_path"`
	Verbose         string `env:"verbose,opt[true,false]"`
	IsCacheDisabled string `env:"is_cache_disabled,opt[true,false]"`
}

func createConfigsModelFromEnvs() (ConfigsModel, error) {
	var cfg ConfigsModel
	if err := stepconf.Parse(&cfg); err != nil {
		return ConfigsModel{}, err
	}

	if cfg.PodfilePath != "" {
		if exist, err := pathutil.IsPathExists(cfg.PodfilePath); err != nil {
			return ConfigsModel{}, fmt.Errorf("failed to check if PodfilePath exists at %s: %s", cfg.PodfilePath, err)
		} else if !exist {
			return ConfigsModel{}, fmt.Errorf("Podfile does not exist at: %s", cfg.PodfilePath)
		}
	}

	return cfg, nil
}
