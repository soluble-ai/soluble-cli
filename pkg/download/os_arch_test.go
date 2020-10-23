package download

import "testing"

func TestOsArch(t *testing.T) {
	// See https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63
	// for a list of goarch/goos values
	testCases := []struct {
		r, o, a string
	}{
		{"soluble_v0.4.19_darwin_amd64.tar.gz", "darwin", "amd64"},
		{"soluble_v0.4.19_linux_386.tar.gz", "linux", "386"},
		{"trivy_0.12.0_Linux-64bit.tar.gz", "linux", "amd64"},
		{"trivy_0.12.0_macOS-64bit.tar.gz", "darwin", "amd64"},
		{"terrascan_1.1.0_Linux_i386.tar.gz", "linux", "386"},
		{"terrascan_1.1.0_Darwin_x86_64.tar.gz", "darwin", "amd64"},
	}

	for _, tc := range testCases {
		if !isMatchingReleaseName(tc.r, tc.o, tc.a) {
			t.Error("failed detection", tc.r, tc.o, tc.a)
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
