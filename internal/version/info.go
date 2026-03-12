package version

import "runtime/debug"

const RepoURL = "https://github.com/chengliang4810/jimuqu-devops.git"

func Current() string {
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
		return info.Main.Version
	}

	return "dev"
}
