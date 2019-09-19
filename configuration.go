package main

import (
	"time"
)

type configuration struct {
	AppRoot            string        `yaml:"app_root" json:"app_root"`
	IgnoredFolders     []string      `yaml:"ignored_folders" json:"ignored_folders"`
	IncludedExtensions []string      `yaml:"included_extensions" json:"included_extensions"`
	BuildDelay         time.Duration `yaml:"build_delay" json:"build_delay"`
}
