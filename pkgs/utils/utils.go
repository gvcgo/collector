package utils

import "strings"

const (
	Windows string = "windows"
	MacOS   string = "darwin"
	Linux   string = "linux"
	X64     string = "amd64"
)

var ArchOSs map[string]string = map[string]string{
	"x86-64":     "amd64",
	"win64":      "amd64",
	"linux64":    "amd64",
	"x86":        "386",
	"i386":       "386",
	"i686":       "386",
	"arm64":      "arm64",
	"armv6":      "arm",
	"ppc64le":    "ppc64le",
	"macos":      "darwin",
	"os x 10.8+": "darwin",
	"os x 10.6+": "darwin",
	"linux":      "linux",
	"windows":    "windows",
	"freebsd":    "freebsd",
}

var ArchMap = map[string]string{
	"amd64":     "amd64",
	"x86-64":    "amd64",
	"x86_64":    "amd64",
	"x64":       "amd64",
	"win64":     "amd64",
	"64-bit":    "amd64",
	"x86":       "386",
	"i586":      "386",
	"i686":      "386",
	"i386":      "386",
	"win32":     "386",
	"32-bit":    "386",
	"arm64":     "arm64",
	"aarch64":   "arm64",
	"aarch_64":  "arm64",
	"arm32":     "arm",
	"armv6":     "arm",
	"ppc64le":   "ppc64le",
	"ppcle_64":  "ppc64le",
	"s390x":     "s390x",
	"s390_64":   "s390x",
	"powerpc64": "ppc64",
	"ppc64":     "ppc64",
	"universal": "universal",
}

var PlatformMap = map[string]string{
	"macosx":  MacOS,
	"apple":   MacOS,
	"darwin":  MacOS,
	"macos":   MacOS,
	"mac":     MacOS,
	"winnt":   Windows,
	"win":     Windows,
	"osx":     MacOS,
	"linux":   Linux,
	"windows": Windows,
	"freebsd": "freebsd",
	"aix":     "aix",
}

func MapArchAndOS(ArchOrOS string) (result string) {
	result, ok := ArchOSs[strings.ToLower(ArchOrOS)]
	if !ok {
		result = ArchOrOS
	}
	return
}

const (
	Win        string = "win"
	Zsh        string = "zsh"
	Bash       string = "bash"
	PowerShell string = "powershell"
)

func ParseArch(name string) string {
	name = strings.ToLower(name)
	if strings.Contains(name, "win32-x64") {
		return "any"
	}
	for k, v := range ArchMap {
		if k == "x86" && strings.Contains(name, k) && (strings.Contains(name, "x86_64") || strings.Contains(name, "x86-64")) {
			continue
		}
		if strings.Contains(name, k) {
			return v
		}
	}
	return ""
}

func ParsePlatform(name string) string {
	name = strings.ToLower(name)
	for k, v := range PlatformMap {
		if k == "win" && strings.Contains(name, k) && strings.Contains(name, "darwin") {
			continue
		} else if strings.Contains(name, k) {
			return v
		}
	}
	return ""
}
