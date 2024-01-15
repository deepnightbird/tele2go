package main

import "runtime"

var version string = "2.9.0 " + runtime.GOOS

// build flags
var (
	BuildTime  string
	CommitHash string
	GoVersion  string
	GitTag     string
)
