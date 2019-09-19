package main

import (
	"time"
)

type configuration struct {
	AppRoot            string
	IgnoredFolders     []string
	IncludedExtensions []string
	BuildDelay         time.Duration
}
