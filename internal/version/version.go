package version

import (
	"fmt"
	"runtime"
)

// Build information. These variables are set via ldflags during build.
var (
	// Version is the current version of the application
	Version = "dev"
	
	// Commit is the git commit hash
	Commit = "unknown"
	
	// Date is the build date
	Date = "unknown"
	
	// Author information
	Author = "Aaron Bockelie"
	
	// Repository URL
	Repository = "https://github.com/aaronsb/atlassian-assets"
)

// Info holds version and build information
type Info struct {
	Version    string `json:"version"`
	Commit     string `json:"commit"`
	Date       string `json:"date"`
	Author     string `json:"author"`
	Repository string `json:"repository"`
	GoVersion  string `json:"go_version"`
	Platform   string `json:"platform"`
}

// GetInfo returns version and build information
func GetInfo() Info {
	return Info{
		Version:    Version,
		Commit:     Commit,
		Date:       Date,
		Author:     Author,
		Repository: Repository,
		GoVersion:  runtime.Version(),
		Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a formatted version string
func (i Info) String() string {
	return fmt.Sprintf("atlassian-assets version %s\ncommit: %s\nbuilt: %s\ngo version: %s\nplatform: %s\nauthor: %s",
		i.Version, i.Commit, i.Date, i.GoVersion, i.Platform, i.Author)
}

// Short returns a short version string
func (i Info) Short() string {
	commit := i.Commit
	if len(commit) > 8 {
		commit = commit[:8]
	}
	return fmt.Sprintf("atlassian-assets %s (%s)", i.Version, commit)
}