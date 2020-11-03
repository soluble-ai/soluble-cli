package download

import (
	"testing"

	"github.com/google/go-github/v32/github"
)

func TestOsArch(t *testing.T) {
	// See https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63
	// for a list of goarch/goos values
	testCases := []struct {
		r, o, a string
		m       bool
	}{
		{"soluble_v0.4.19_darwin_amd64.tar.gz", "darwin", "amd64", true},
		{"soluble_v0.4.19_linux_386.tar.gz", "linux", "386", true},
		{"trivy_0.12.0_Linux-64bit.tar.gz", "linux", "amd64", true},
		{"trivy_0.12.0_macOS-64bit.tar.gz", "darwin", "amd64", true},
		{"terrascan_1.1.0_Linux_i386.tar.gz", "linux", "386", true},
		{"terrascan_1.1.0_Darwin_x86_64.tar.gz", "darwin", "amd64", true},
		{"tfsec-checkgen-darwin-amd64", "darwin", "amd64", false},
		{"tfsec-darwin-amd64", "darwin", "amd64", true},
	}

	for _, tc := range testCases {
		if tc.m {
			if !isMatchingReleaseName(tc.r, tc.o, tc.a) {
				t.Error("failed detection", tc.r, tc.o, tc.a)
			}
		}
	}
	for _, a := range []string{"amd64", "386"} {
		for _, o := range []string{"linux", "darwin"} {
			for _, tc := range testCases {
				if tc.a != a && tc.o != o {
					if isMatchingReleaseName(tc.r, a, o) {
						t.Error("false detection", tc.r, a, o)
					}
				}
			}
		}
	}
}

func TestReleasePriority(t *testing.T) {
	assets := []*github.ReleaseAsset{
		makeReleaseAsset("foo.tar.gz"),
		makeReleaseAsset("foo.zip"),
		makeReleaseAsset("foo.deb"),
		makeReleaseAsset("foo"),
	}
	asset, err := chooseReleaseAsset(assets, DefaultReleasePriority)
	if asset == nil || err != nil {
		t.Error(asset, err)
	}
	if asset.GetName() != "foo.tar.gz" {
		t.Error(asset)
	}
	_, err = chooseReleaseAsset(assets, func(string) ReleasePriority { return Match })
	if err == nil {
		t.Error("should have errored")
	}
}

func makeReleaseAsset(name string) *github.ReleaseAsset {
	return &github.ReleaseAsset{Name: &name}
}
