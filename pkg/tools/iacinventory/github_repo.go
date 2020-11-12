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
	CISystems []string `json:"ci_systems,omitempty"`

	// TerraformDirs are directories that contain '.tf' files
	TerraformDirs []string `json:"terraform_dirs,omitempty"`

	// CloudformationDirs are directories that contain cloudformation files
	CloudformationDirs []string `json:"cloudformation_dirs,omitempty"`

	// DockerfileDirs are... Dockerfile files.
	DockerfileDirs []string `json:"dockerfile_files,omitempty"`

	// K8sManifestDirs are directories that contain Kubernetes manifest files
	K8sManifestDirs []string `json:"k8s_manifest_dirs,omitempty"`
}

// getRepos fetches the list of all repositories accessiable to a given credential pair.
func getRepos(oauthToken string, all bool, repoNames []string) ([]*github.Repository, error) {
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
		for {
			pageCtx, cancel := context.WithTimeout(ctx, time.Second*10)
			defer cancel()
			pageRepos, resp, err := client.Repositories.List(pageCtx, "", opt)
			defer resp.Body.Close()
			if err != nil {
				return nil, err
			}
			repos = append(repos, pageRepos...)
			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
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
func (g *GithubRepo) downloadTarball(oauthToken string, repo *github.Repository, dir string) (string, error) {
	if oauthToken == "" {
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

// we need to differentiate github vs local scans, and tmpdir
// consistency is the easiest way to fix it.
const githubTmpDirPrefix string = "github-iac-inventory-tmpdir"

func (g *GithubRepo) downloadAndScan(token string, repo *github.Repository) error {
	dir, err := ioutil.TempDir("", githubTmpDirPrefix+"*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	tarball, err := g.downloadTarball(token, repo, dir)
	if err != nil {
		return fmt.Errorf("could not download and unpack %s: %w", repo.GetFullName(), err)
	}
	return g.scanTarball(dir, tarball)
}

func (g *GithubRepo) scanTarball(dir, tarball string) error {
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

	// Github packages files into a sub-directory with a repo-relative prefix
	// ex: ./soluble-ai-fizzbuzz-999999999999999999999/
	unarchivedDirPath := dir
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != unarchivedDirPath {
			unarchivedDirPath = path
			return os.ErrExist
		}
		return nil
	}); err != nil && err != os.ErrExist {
		return fmt.Errorf("error determining GitHub tarball root")
	}
	if unarchivedDirPath == "" {
		return fmt.Errorf("could not find unarchived repository")
	}
	res, err := Directory(unarchivedDirPath)
	if err != nil {
		return fmt.Errorf("error scanning directory: %w", err)
	}

	g.CISystems = res.CISystems
	g.TerraformDirs = res.TerraformDirs
	g.CloudformationDirs = res.CloudformationDirs
	g.DockerfileDirs = res.DockerfileDirs
	g.K8sManifestDirs = res.K8sManifestDirs

	log.Infof("{primary:%s} has {info:%d} terraform directories, {info:%d} cloudformation directories, {info:%d} dockerfiles, {info:%d} k8s manifest directories",
		g.FullName, len(g.TerraformDirs), len(g.CloudformationDirs), len(g.DockerfileDirs), len(g.K8sManifestDirs))

	return nil
}

type Results struct {
	CISystems []string `json:"ci_systems"`
	// CIFiles             []string `json:"ci_files"` // TBD
	TerraformFiles      []string `json:"terraform_files"`
	TerraformDirs       []string `json:"terraform_dirs"`
	CloudformationFiles []string `json:"cloudformation_files"`
	CloudformationDirs  []string `json:"cloudformation_dirs"`
	DockerfileFiles     []string `json:"dockerfile_files"`
	DockerfileDirs      []string `json:"dockerfile_dirs"`
	K8sManifestFiles    []string `json:"k_8_s_manifest_files"`
	K8sManifestDirs     []string `json:"k_8_s_manifest_dirs"`
}

func Directory(dir string) (*Results, error) {
	ciSystems := util.NewStringSet()
	terraformFiles := util.NewStringSet()
	terraformDirs := util.NewStringSet()
	cloudformationFiles := util.NewStringSet()
	cloudformationDirs := util.NewStringSet()
	dockerfileFiles := util.NewStringSet()
	dockerfileDirs := util.NewStringSet()
	k8sManifestFiles := util.NewStringSet()
	k8sManifestDirs := util.NewStringSet()

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		ci, err := walkCI(path, info, err)
		if err != nil {
			return err
		}
		if ci != "" {
			ciSystems.Add(string(ci))
		}
		if info.Mode().IsRegular() {
			fileName, _ := filepath.Rel(dir, path)
			// handle GitHub tmpdir
			splitPath := strings.SplitN(fileName, string(filepath.Separator), 2)
			pathRelativeToRepoRoot := "." // include the root directory
			if len(splitPath) > 1 {
				// if we're not in the root, set the directory
				// relative to the git repository root
				pathRelativeToRepoRoot = splitPath[1]
			}
			// handle local paths
			if !strings.Contains(path, githubTmpDirPrefix) {
				pathRelativeToRepoRoot = fileName // there is no "repo root" here.
			}
			if isTerraformFile(path, info) {
				terraformFiles.Add(pathRelativeToRepoRoot)
				terraformDirs.Add(filepath.Dir(pathRelativeToRepoRoot))
			}
			if isCloudFormationFile(path, info) {
				cloudformationFiles.Add(pathRelativeToRepoRoot)
				cloudformationDirs.Add(filepath.Dir(pathRelativeToRepoRoot))
			}
			if isDockerFile(path, info) {
				dockerfileFiles.Add(pathRelativeToRepoRoot)
				dockerfileDirs.Add(filepath.Dir(pathRelativeToRepoRoot))
			}
			if isKubernetesManifest(path, info) {
				k8sManifestFiles.Add(pathRelativeToRepoRoot)
				k8sManifestDirs.Add(filepath.Dir(pathRelativeToRepoRoot))
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Results{
		CISystems:           ciSystems.Values(),
		TerraformFiles:      terraformFiles.Values(),
		TerraformDirs:       terraformDirs.Values(),
		CloudformationFiles: cloudformationFiles.Values(),
		CloudformationDirs:  cloudformationDirs.Values(),
		DockerfileFiles:     dockerfileFiles.Values(),
		DockerfileDirs:      dockerfileDirs.Values(),
		K8sManifestFiles:    k8sManifestFiles.Values(),
		K8sManifestDirs:     k8sManifestDirs.Values(),
	}, nil
}
