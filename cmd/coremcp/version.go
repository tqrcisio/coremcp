package main

// Version information is injected at build time via -ldflags (see Makefile).
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)
