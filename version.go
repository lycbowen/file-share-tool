package fileshare

import "runtime/debug"

const commandName = "fshare"

func VersionString() string {
	version := "dev"
	info, ok := debug.ReadBuildInfo()
	if ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		version = info.Main.Version
	}
	return commandName + " " + version
}
