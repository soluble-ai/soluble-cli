package download

import (
	"runtime"
	"strings"
)

var avoidSubstrings = []string{
	"-checkgen-",
}

var archSubstrings = map[string][]string{
	"amd64": {"_amd64", "_x86_64", "-64bit", "-amd64"},
	"386":   {"_386", "_x86", "_i386", "-32bit"},
}

var osSubstrings = map[string][]string{
	"linux":   {"_linux"},
	"darwin":  {"_darwin", "_macos", "_osx", "-darwin-"},
	"windows": {"_windows", "-windows"},
}

func isThisRuntimeRelease(r string) bool {
	return isMatchingReleaseName(r, runtime.GOOS, runtime.GOARCH)
}

func isMatchingReleaseName(r, o, a string) bool {
	r = strings.ToLower(r)
	archOK := false
	for _, a := range archSubstrings[a] {
		if strings.Contains(r, a) {
			archOK = true
			break
		}
	}
	if !archOK {
		return false
	}
	for _, o := range osSubstrings[o] {
		if strings.Contains(r, o) {
			for _, n := range avoidSubstrings {
				if strings.Contains(r, n) {
					return false
				}
			}
			return true
		}
	}
	return false
}