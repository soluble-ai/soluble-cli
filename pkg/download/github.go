package download

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/google/go-github/v32/github"
)

func isLatestTag(tag string) bool {
	return tag == "" || tag == "latest"
}

func getGithubReleaseAsset(owner, repo, tag string) (*github.RepositoryRelease, *github.ReleaseAsset, error) {
	client := github.NewClient(nil)
	var release *github.RepositoryRelease
	var err error
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()
	if isLatestTag(tag) {
		release, _, err = client.Repositories.GetLatestRelease(ctx, owner, repo)
	} else {
		release, _, err = client.Repositories.GetReleaseByTag(ctx, owner, repo, tag)
	}
	if err != nil {
		return nil, nil, err
	}
	assets, _, err := client.Repositories.ListReleaseAssets(ctx, owner, repo, release.GetID(), nil)
	if err != nil {
		return nil, nil, err
	}
	for _, asset := range assets {
		name := asset.GetName()
		if isThisRuntimeRelease(name) {
			return release, asset, nil
		}
	}
	return nil, nil, fmt.Errorf("cannot find release for %s/%s", owner, repo)
}

var archNames = map[string][]string{
	"amd64": {"x86_64"},
	"386":   {"x86", "i386"},
	"arm64": {"arm64"},
	"arm":   {"arm"},
}

var osNames = map[string][]string{
	"darwin": {"macos", "osx"},
	"linux":  {"linux"},
}

func isThisRuntimeRelease(r string) bool {
	r = strings.ToLower(r)
	if strings.Contains(r, runtime.GOOS) {
		return true
	}
	oses := osNames[runtime.GOOS]
	for _, os := range oses {
		if strings.Contains(r, os) {
			return true
		}
	}
	if strings.Contains(r, runtime.GOARCH) {
		return true
	}
	names := archNames[runtime.GOARCH]
	for _, name := range names {
		if strings.Contains(r, name) {
			return true
		}
	}
	return false
}
