package terraform

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"runtime"
	"sort"

	"github.com/hashicorp/go-version"
)

var aPattern = regexp.MustCompile(`<a href="[^"]*">terraform_(.*)</a>`)

func GetVersionAndURL(requestedVersion string) (version string, url string, err error) {
	version = requestedVersion
	if version == "" || version == "latest" {
		var resp *http.Response
		resp, err = http.Get("https://releases.hashicorp.com/terraform/")
		if err != nil {
			return
		}
		defer resp.Body.Close()
		version = parseLatestVersion(resp.Body)
	}
	url = fmt.Sprintf("https://releases.hashicorp.com/terraform/%s/terraform_%s_%s_%s.zip",
		version, version, runtime.GOOS, runtime.GOARCH)
	return
}

func parseLatestVersion(r io.Reader) string {
	scanner := bufio.NewScanner(r)
	var versions []*version.Version
	for scanner.Scan() {
		m := aPattern.FindStringSubmatch(scanner.Text())
		if m != nil {
			v, err := version.NewVersion(m[1])
			if err == nil && v.Prerelease() == "" {
				versions = append(versions, v)
			}
		}
	}
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].LessThan(versions[j])
	})
	if len(versions) > 0 {
		return versions[len(versions)-1].Original()
	}
	return ""
}
