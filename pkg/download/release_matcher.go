package download

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/google/go-github/v32/github"
)

type ReleasePriority int

const (
	NoMatch ReleasePriority = 0
	Match   ReleasePriority = 100
)

type GithubReleaseMatcher func(string) ReleasePriority

func chooseReleaseAsset(assets []*github.ReleaseAsset, m GithubReleaseMatcher) (*github.ReleaseAsset, error) {
	if m == nil {
		m = DefaultReleaseMatcher
	}
	var highestPriorityAsset *github.ReleaseAsset
	priority := NoMatch
	for _, asset := range assets {
		if p := m(asset.GetName()); p != NoMatch {
			if highestPriorityAsset == nil {
				highestPriorityAsset = asset
				priority = p
			} else if priority == p {
				return nil, fmt.Errorf("multiple matching assets found: %s and %s", highestPriorityAsset.GetName(),
					asset.GetName())
			}
		}
	}
	if highestPriorityAsset != nil {
		return highestPriorityAsset, nil
	}
	names := []string{}
	for _, asset := range assets {
		names = append(names, asset.GetName())
	}
	return nil, fmt.Errorf("could not find a matching release asset from: %s", strings.Join(names, " "))
}

var avoidSubstrings = []string{
	"-checkgen-",
}

var archSubstrings = map[string][]string{
	"amd64": {"_amd64", "_x86_64", "-64bit", "-amd64"},
	"386":   {"_386", "_x86", "_i386", "-32bit"},
}

var osSubstrings = map[string][]string{
	"linux":   {"_linux", "-linux-"},
	"darwin":  {"_darwin", "_macos", "_osx", "-darwin-", "-osx-"},
	"windows": {"_windows", "-windows"},
}

func DefaultReleaseMatcher(r string) ReleasePriority {
	if isMatchingReleaseName(r, runtime.GOOS, runtime.GOARCH) {
		return DefaultReleasePriority(r)
	}
	return NoMatch
}

func DefaultReleasePriority(r string) ReleasePriority {
	r = strings.ToLower(r)
	switch {
	case strings.HasSuffix(r, ".tar.gz"):
		return ReleasePriority(100)
	case strings.HasSuffix(r, ".zip"):
		return ReleasePriority(99)
	case strings.HasSuffix(r, ".deb") || strings.HasSuffix(r, ".rpm") || strings.HasSuffix(r, ".apk"):
		return NoMatch
	default:
		return ReleasePriority(1)
	}
}

func isMatchingReleaseName(r, o, a string) bool {
	return IsMatchingArch(r, a) && IsMatchingOS(r, o)
}

func IsMatchingArch(r, a string) bool {
	r = strings.ToLower(r)
	for _, a := range archSubstrings[a] {
		if strings.Contains(r, a) {
			return true
		}
	}
	return false
}

func IsMatchingOS(r, o string) bool {
	r = strings.ToLower(r)
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
