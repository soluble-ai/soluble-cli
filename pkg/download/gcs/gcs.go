package gcs

import (
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
)

type GCSResolver struct {
	bucket string
	name   string
}

func NewResolver(bucket, name string) func(string) (version string, url string, err error) {
	gcs := &GCSResolver{
		bucket: bucket,
		name:   name,
	}
	return gcs.GetVersionAndURL
}

func (gcs *GCSResolver) GetVersionAndURL(requestedVersion string) (version string, url string, err error) {
	if requestedVersion == "" || requestedVersion == "latest" {
		version, err = gcs.findLatestVersion()
		if err != nil {
			return
		}
	} else {
		version = requestedVersion
	}
	ersion := version[1:]
	url = fmt.Sprintf("https://storage.googleapis.com/storage/v1/b/%s/o/%s%%2F%s%%2F%s_%s_%s_%s.tar.gz?alt=media",
		gcs.bucket, gcs.name, version, gcs.name, ersion, runtime.GOOS, runtime.GOARCH)
	return
}

func (gcs *GCSResolver) findLatestVersion() (string, error) {
	url := fmt.Sprintf("https://storage.googleapis.com/storage/v1/b/%s/o/%s%%2Flatest.txt?alt=media",
		gcs.bucket, gcs.name)
	// #nosec G107
	resp, err := http.Get(url)
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
