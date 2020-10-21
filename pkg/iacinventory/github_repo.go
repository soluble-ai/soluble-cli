package iacinventory

import (
	"archive/tar"
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
	"time"

	"github.com/google/go-github/github"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"golang.org/x/oauth2"
)

var _ Repo = &GithubRepo{}

type GithubRepo struct {
	// FullName includes the organization/user, as in "soluble-ai/example".
	FullName string `json:"full_name"`
	// Name is only the repository name, as in "example".
	Name string `json:"name"`

	// CI is the repo's configured CI system, if present.
	CI []CI

	// TerraformDirs are directories that contain '.tf' files
	TerraformDirs []string

	// CloudformationDirs are directories that contain cloudformation files
	CloudformationDirs []string

	// dir is where we've unarchived the tarball for analysis.
	dir string

	repo *github.Repository
}

func (g GithubRepo) getName() string {
	return g.Name
}

func (g GithubRepo) getCISystems() []string {
	var out []string
	for _, ci := range g.CI {
		out = append(out, string(ci))
	}
	return out
}

func (g GithubRepo) getTerraformDirs() []string {
	var out []string
	for _, v := range g.TerraformDirs {
		out = append(out, strings.TrimPrefix(filepath.Dir(filepath.Clean(v)), g.dir+"/"))
	}
	return out
}

func (g GithubRepo) getCloudformationDirs() []string {
	var out []string
	for _, v := range g.TerraformDirs {
		out = append(out, strings.TrimPrefix(filepath.Dir(filepath.Clean(v)), g.dir+"/"))
	}
	return out
}

// getRepos fetches the list of all repositories accessiable to a given credential pair.
func getRepos(username, oauthToken string) ([]*github.Repository, error) {
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

// download the tarball of the default branch for a repository.
func (g *GithubRepo) download(username, oauthToken string, repo *github.Repository) ([]byte, error) {
	if username == "" || oauthToken == "" {
		return nil, fmt.Errorf("error fetching repository: credentials are not set")
	}
	ctx := context.TODO()
	url := "https://api.github.com/repos/" + g.FullName + "/tarball"
	log.Debugf("downloading repository tarball at: %s", url)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("Authorization", "token "+oauthToken)
	req.Header.Set("Accept", "application/vnd.github.v3.raw")
	c := http.Client{
		Timeout: time.Second * 120, // sometimes pulling github repos can be slow (on GitHub's end).
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending HTTP request: %w", err)
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// extract tarballs of a github repository to g.dir
func (g *GithubRepo) extract(r io.Reader) error {
	// we stuff the extracted tarball into a temporary directory
	dir, err := ioutil.TempDir("", "soluble-iac-scan")
	if err != nil {
		return fmt.Errorf("error creating temporary directory: %w", err)
	}
	g.dir = dir
	gunzip, err := gzip.NewReader(r)
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
		log.Debugf("writing file: %s", outfile)
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
					return fmt.Errorf("file in tarball for repository %q was larger than 1MB - skipping", g.FullName)
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

// getCode fetches and extracts the tarball of the master branch.
func (g *GithubRepo) getCode(username string, oauthToken string) error {
	log.Debugf("downloading tarball for repo: %s", g.FullName)
	tarballData, err := g.download(username, oauthToken, g.repo)
	if err != nil {
		return err
	}
	log.Debugf("downloaded tarball for repo: %s", g.FullName)
	r := bytes.NewReader(tarballData)
	log.Debugf("extracting tarball for repo: %s", g.FullName)
	if err := g.extract(r); err != nil {
		return err
	}
	log.Debugf("extracted tarball for repo: %s", g.FullName)
	// TODO: validate that files actually ended up in g.dir
	return nil
}
