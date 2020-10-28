package iacinventory

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/soluble-ai/soluble-cli/pkg/archive"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"golang.org/x/oauth2"
)

type GithubRepo struct {
	// FullName includes the organization/user, as in "soluble-ai/example".
	FullName string `json:"full_name"`
	// Name is only the repository name, as in "example".
	Name string `json:"name"`

	// "github.com/"+FullName
	GitRepo string `json:"git_repo"`

	// CI is the repo's configured CI system, if present.
	CISystems []CI `json:"ci_systems,omitempty"`

	// TerraformDirs are directories that contain '.tf' files
	TerraformDirs []string `json:"terraform_dirs,omitempty"`

	// CloudformationDirs are directories that contain cloudformation files
	CloudformationDirs []string `json:"cloudformation_dirs,omitempty"`

	// DockerfileFiles are... Dockerfile files.
	DockerfileFiles []string `json:"dockerfile_files,omitempty"`

	// K8sManifestDirs are directories that contain Kubernetes manifest files
	K8sManifestDirs []string `json:"k8s_manifest_dirs,omitempty"`
}

// getRepos fetches the list of all repositories accessiable to a given credential pair.
func getRepos(username, oauthToken string, all bool, repoNames []string) ([]*github.Repository, error) {
	ctx, cf := context.WithTimeout(context.Background(), 30*time.Second)
	defer cf()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: oauthToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	var repos []*github.Repository
	if len(repoNames) == 0 {
		opt := &github.RepositoryListOptions{
			ListOptions: github.ListOptions{PerPage: 10}, // TODO: revert to 100
		}
		if !all {
			opt.Visibility = "public"
		}
		var err error
		repos, _, err = client.Repositories.List(ctx, "", opt)
		if err != nil {
			return nil, err
		}
	} else {
		const githubcom = "github.com/"
		for _, name := range repoNames {
			n := name
			if strings.HasPrefix(name, githubcom) {
				n = n[len(githubcom):]
			}
			p := strings.Split(n, "/")
			if len(p) != 2 {
				return nil, fmt.Errorf("illegal github repository name %s", name)
			}
			repo, _, err := client.Repositories.Get(ctx, p[0], p[1])
			if err != nil {
				return nil, err
			}
			repos = append(repos, repo)
		}
	}
	return repos, nil
}

// download the tarball of the default branch for a repository.
func (g *GithubRepo) downloadTarball(username, oauthToken string, repo *github.Repository, dir string) (string, error) {
	if username == "" || oauthToken == "" {
		return "", fmt.Errorf("error fetching repository: credentials are not set")
	}
	url := "https://api.github.com/repos/" + g.FullName + "/tarball"
	log.Infof("Downloading {primary:%s}", url)
	ctx, cf := context.WithTimeout(context.Background(), time.Second*120)
	defer cf()
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("Authorization", "token "+oauthToken)
	req.Header.Set("Accept", "application/vnd.github.v3.raw")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not download tarball: %w", err)
	}
	defer resp.Body.Close()
	f, err := os.Create(filepath.Join(dir, ".tarball.tar.gz"))
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return "", err
	}
	return f.Name(), nil
}

func (g *GithubRepo) downloadAndScan(user, token string, repo *github.Repository) error {
	dir, err := ioutil.TempDir("", "iacinventory*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	tarball, err := g.downloadTarball(user, token, repo, dir)
	if err != nil {
		return fmt.Errorf("could not download and unpack %s: %w", repo.GetFullName(), err)
	}
	info, err := os.Stat(tarball)
	if err == nil {
		log.Infof("Tarball is {info:%dK}", info.Size()>>10)
	}
	err = archive.Do(archive.Untar, tarball, dir, &archive.Options{
		TruncateFileSize: 1 << 20,
		IgnoreSymLinks:   true,
	})
	if err != nil {
		return fmt.Errorf("could not unpack tarball: %w", err)
	}

	terraformDirs := util.NewStringSet()
	cfnDirs := util.NewStringSet()
	dockerFiles := util.NewStringSet()
	k8sManifestDirs := util.NewStringSet()
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		ci, err := walkCI(path, info, err)
		if ci != "" {
			contains := false
			for _, cisys := range g.CISystems {
				if cisys == ci {
					contains = true
				}
			}
			if !contains {
				g.CISystems = append(g.CISystems, ci)
			}
		}
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			dirName, _ := filepath.Rel(dir, filepath.Dir(path))
			splitPath := strings.SplitN(dirName, string(filepath.Separator), 2)
			if len(splitPath) == 1 {
				// walk on the root folder
				return nil
			}
			pathRelativeToRepoRoot := splitPath[1]
			if isTerraformFile(path, info) {
				terraformDirs.Add(pathRelativeToRepoRoot)
			}
			if isCloudFormationFile(path, info) {
				cfnDirs.Add(pathRelativeToRepoRoot)
			}
			if isDockerFile(path, info) {
				dockerFiles.Add(pathRelativeToRepoRoot)
			}
			if isKubernetesManifest(path, info) {
				k8sManifestDirs.Add(pathRelativeToRepoRoot)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	g.TerraformDirs = terraformDirs.Values()
	g.CloudformationDirs = cfnDirs.Values()
	g.DockerfileFiles = dockerFiles.Values()
	g.K8sManifestDirs = k8sManifestDirs.Values()
	log.Infof("{primary:%s} has {info:%d} terraform directories, {info:%d} cloudformation directories, {info:%d} dockerfiles, {info:%d} k8s manifest directories",
		g.FullName, len(g.TerraformDirs), len(g.CloudformationDirs), len(g.DockerfileFiles), len(g.K8sManifestDirs))
	return nil
}
