package iacinventory

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/log"
)

type GithubRepo struct {
	// FullName includes the organization/user, as in "soluble-ai/example".
	FullName string `json:"full_name"`
	// Name is only the repository name, as in "example".
	Name string `json:"name"`

	// CI is the repo's configured CI system, if present.
	CI []CI

	// TerraformDirs are directories that contain '.tf' files
	TerraformDirs map[string]bool // DEBT: lazy map implementation

	// CloudformationDirs are directories that contain cloudformation files
	CloudformationDirs map[string]bool // DEBT: lazy map implementation

	// dir is where we've unarchived the tarball for analysis.
	dir string
}

// getTerraformDirs implements WalkFunc to search for directories that contain Terraform files.
func (g *GithubRepo) getTerraformDirs(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	// if it is anything other than a regular file, return
	if !info.Mode().IsRegular() {
		return nil
	}
	// If the file ends with TF, the parent directory contains terraform files
	if strings.HasSuffix(info.Name(), ".tf") {
		if g.TerraformDirs == nil {
			g.TerraformDirs = make(map[string]bool)
		}
		g.TerraformDirs[strings.TrimPrefix(filepath.Dir(filepath.Clean(path)), g.dir+"/")] = true
	}
	return nil
}

// getCloudFormationDirs implements WalkFunc to search for CloudFormation files in a repository.
func (g *GithubRepo) getCloudFormationDirs(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	// if it is anything other than a regular file, return
	if !info.Mode().IsRegular() {
		return nil
	}
	// Cloudformation files do not have a unique extension, and are *typically*
	// ".yaml" or ".json" by convention. However, sometimes organizations use
	// Jinja, Go, or some other utility to template their Cloudformation.
	//
	// If the file has a possible extension and contains the string 'AWSTemplateFormatVersion',
	// then it is *most likely* a CF file.

	const maxSizeForCloudFormationConsideration int64 = 5 << (10 * 2) // 5MB, which is VERY large for json/yaml data
	if info.Size() > maxSizeForCloudFormationConsideration {
		// Exit early - file disqualified due to size.
		return nil
	}

	if strings.HasSuffix(info.Name(), ".json") ||
		strings.HasSuffix(info.Name(), ".yaml") ||
		strings.HasSuffix(info.Name(), ".yml") ||
		strings.HasSuffix(info.Name(), ".template") {
		f, err := os.Open(filepath.Clean(path))
		if err != nil {
			return fmt.Errorf("error opening file during CloudFormation analysis: %w", err)
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if bytes.Contains(scanner.Bytes(), []byte("AWSTemplateFormatVersion")) {
				if g.CloudformationDirs == nil {
					g.CloudformationDirs = make(map[string]bool)
				}
				g.CloudformationDirs[strings.TrimPrefix(filepath.Dir(filepath.Clean(path)), g.dir+"/")] = true
			}
		}
	}
	return nil
}

/*
// fetch the github repos using the GithubAPI
func (g *GithubRepo) fetch(username, oauthToken string) ([]*github.Repository, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: oauthToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	opt := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	repos, _, err := client.Repositories.List(ctx, "", opt)
	if err != nil {
		return nil, err
	}
	return repos, err
}

// extract downloads and extracts tarballs of the github repository
func (g *GithubRepo) extract(repo *github.Repository) error {
	url := r.GetArchiveURL()
	// we stuff the extracted tarball into a temporary directory
	dir, err := ioutil.TempDir("", "soluble-iac-scan")
	if err != nil {
		return fmt.Errorf("error creating temporary directory: %w", err)
	}
	g.dir = dir
	return nil
}
*/

// getMaster fetches and extracts the tarball of the master branch.
func (g *GithubRepo) getMaster(ctx context.Context, c *http.Client, username string, oauthToken string) error {
	if username == "" || oauthToken == "" {
		return fmt.Errorf("error fetching repository: credentials are not set")
	}
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/repos/"+g.FullName+"/tarball/master", nil)
	req.Header.Set("Authorization", "token "+oauthToken)
	req.Header.Set("Accept", "application/vnd.github.v3.raw")
	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("error sending HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// we stuff the extracted tarball into a temporary directory
	dir, err := ioutil.TempDir("", "soluble-iac-scan")
	if err != nil {
		return fmt.Errorf("error creating temporary directory: %w", err)
	}
	g.dir = dir
	gunzip, err := gzip.NewReader(resp.Body)
	if err != nil {
		// If there is no repository data, we'll encounter a gzip error
		if errors.Is(err, gzip.ErrHeader) {
			return fmt.Errorf("repository has no contents")
		}
		return fmt.Errorf("error creating gzip reader for tarball: %w", err)
	}
	defer gunzip.Close()
	t := tar.NewReader(gunzip)

	// write the tarball'd files to disk in the temporary directory
	for {
		header, err := t.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Errorf("error extracting file from tarball: %w", err)
			break
		}

		// cut out the top level directory to place files directly in g.dir
		outfile := filepath.Join(g.dir,
			filepath.Clean(strings.Join(strings.Split(header.Name, string(filepath.Separator))[1:], string(filepath.Separator))))
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(outfile, 0o755); err != nil {
				if os.IsExist(err) {
					continue
				}
				return fmt.Errorf("error creating directory for tarball extraction: %w", err)
			}
		case tar.TypeReg:
			// cut out the outer directory from the filename to ensure we add the files to g.dir (and not g.dir/some-random-subdir)
			err := func() error {
				f, err := os.Create(outfile)
				if err != nil {
					return fmt.Errorf("error creating file: %w", err)
				}
				defer f.Close()
				c, err := io.CopyN(f, t, 1<<(10*2)) // 1MB max size
				if err != nil {
					if !errors.Is(err, io.EOF) {
						return fmt.Errorf("eror copying file: %w", err)
					}
				}
				if c == 1<<(10*2) { // 1MB max file size
					return fmt.Errorf("repository %q was larger than 100MB - skipping", g.FullName)
				}
				return nil
			}()
			if err != nil {
				log.Infof("error extracting tarball: %v", err)
				continue
			}
		case tar.TypeXGlobalHeader, tar.TypeSymlink, tar.TypeLink:
			// we silently ignore (sym)links without failure, as they are not relevant (, as we walk all files)
			continue
		default:
			return fmt.Errorf("unknown filetype %q encountered during extraction of %q archive",
				header.Typeflag, header.Name)
		}
	}
	return nil
}
