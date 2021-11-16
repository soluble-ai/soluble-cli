package tfscore

import (
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
)

func GetVersionAndURL(requestedVersion string) (version string, url string, err error) {
	if requestedVersion == "" || requestedVersion == "latest" {
		version, err = findLatestVersion()
		if err != nil {
			return
		}
	} else {
		version = requestedVersion
	}
	ersion := version[1:]
	url = fmt.Sprintf("https://storage.googleapis.com/storage/v1/b/soluble-public/o/tfscore%%2F%s%%2Ftfscore_%s_%s_%s.tar.gz?alt=media",
		version, ersion, runtime.GOOS, runtime.GOARCH)
	return
}

func findLatestVersion() (string, error) {
	resp, err := http.Get("https://storage.googleapis.com/storage/v1/b/soluble-public/o/tfscore%2Flatest.txt?alt=media")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	dat, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(dat)), nil
}
