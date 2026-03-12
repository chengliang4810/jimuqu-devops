package version

import (
	"runtime/debug"
	"strings"
)

const (
	RepoURL = "https://github.com/chengliang4810/jimuqu-devops.git"
	AppName = "jimuqu-devops"
)

var (
	Version   = "dev"
	Commit    = ""
	BuildTime = ""
)

func Current() string {
	if Version != "" && Version != "dev" {
		return Normalize(Version)
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}

	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" && setting.Value != "" {
			if len(setting.Value) > 8 {
				return setting.Value[:8]
			}
			return setting.Value
		}
	}

	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		return Normalize(info.Main.Version)
	}

	return "dev"
}

func Normalize(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 0 && value[0] == 'v' {
		return value[1:]
	}
	return value
}
