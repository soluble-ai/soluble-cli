package terraform

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"runtime"
	"sort"
	"strconv"
)

var aPattern = regexp.MustCompile(`<a href="[^"]*">terraform_(.*)</a>`)
var vPattern = regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)$`)

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
	var versions []string
	for scanner.Scan() {
		m := aPattern.FindStringSubmatch(scanner.Text())
		if m != nil {
			versions = append(versions, m[1])
		}
	}
	sort.Slice(versions, func(i, j int) bool {
		imajor, iminor, ipatch := parseVersion(versions[i])
		jmajor, jminor, jpatch := parseVersion(versions[j])
		switch {
		case imajor < jmajor:
			return true
		case imajor > jmajor:
			return false
		case iminor < jminor:
			return true
		case iminor > jminor:
			return false
		case ipatch < jpatch:
			return true
		case ipatch > jpatch:
			return false
		}
		return false
	})
	if len(versions) > 0 {
		return versions[len(versions)-1]
	}
	return ""
}

func parseVersion(v string) (major, minor, patch int) {
	m := vPattern.FindStringSubmatch(v)
	if m != nil {
		major, _ = strconv.Atoi(m[1])
		minor, _ = strconv.Atoi(m[2])
		patch, _ = strconv.Atoi(m[3])
	}
	return
}
