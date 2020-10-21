package iacinventory

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"gopkg.in/yaml.v2"
)

var _ IacInventorier = &GithubInventorier{}

type GithubInventorier struct {
	User       string `yaml:"user"`
	OauthToken string `yaml:"oauth_token"`

	repos map[string]GithubRepo
}

// getRepos fetches (and expands) the repositories associated with a github account
func (g *GithubInventorier) getRepos() error {
	if g.User == "" || g.OauthToken == "" {
		return fmt.Errorf("no credentials provided")
	}

	repos, err := getRepos(g.User, g.OauthToken)
	if err != nil {
		return fmt.Errorf("error getting repositories: %w", err)
	}

	var githubRepos []GithubRepo
	for i := range repos {
		githubRepos = append(githubRepos, GithubRepo{
			Name:     *repos[i].Name,
			FullName: *repos[i].FullName,
			repo:     repos[i],
		})
	}

	// reposWithCode are the fully initialized repos, including source code
	reposWithCode := make(map[string]GithubRepo)
	for _, repo := range githubRepos {
		log.Infof("[%s]: analyzing repo...\n", repo.FullName)
		if err := repo.getCode(g.User, g.OauthToken); err != nil {
			log.Infof("[%s]: error fetching archive: %v\n", repo.FullName, err)
		}
		repo := repo // scope pin, an unfortunate go-ism
		// While the walk code below is clean, it is not very optimized or DRY.
		// (though neither is a problem... yet.)
		// Each checker below is left to do its own `scaner.Scan()` of the file. Ideally, any file-content
		// checks should be placed into an asynchronous routine that ingests the rusults of the
		// scanner via a channel.
		if err := filepath.Walk(repo.dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			ci, err := walkCI(path, info, err)
			repo.CI = append(repo.CI, ci)
			if err != nil {
				return err
			}
			tfDir, err := walkTerraformDirs(path, info, err)
			repo.TerraformDirs = append(repo.TerraformDirs, tfDir)
			if err != nil {
				return err
			}
			cfDir, err := walkCloudFormationDirs(path, info, err)
			if err != nil {
				return err
			}
			repo.CloudformationDirs = append(repo.CloudformationDirs, cfDir)
			return nil
		}); err != nil {
			log.Infof("error walking repository %q: %v\n", repo.FullName, err)
		}
		reposWithCode[repo.FullName] = repo
	}
	if len(reposWithCode) == 0 {
		return fmt.Errorf("no repositories contained any code to be analyzed")
	}
	g.repos = reposWithCode
	return nil
}

func (g *GithubInventorier) Run() ([]Repo, error) {
	var err error
	// if no authentication options have been set,
	// we default to reading from the filesystem.
	if g.User == "" && g.OauthToken == "" {
		g.User, g.OauthToken, err = githubCredsFromFS()
		if err != nil {
			return nil, fmt.Errorf("error getting gh credentials from configuration files: %w", err)
		}
		if g.User == "" || g.OauthToken == "" {
			return nil, fmt.Errorf("internal error: no credentials and no error during iacinventory.Run()")
		}
	}

	// fetch the repositories associated with the account
	if err := g.getRepos(); err != nil {
		return nil, fmt.Errorf("error fetching github repositories: %w", err)
	}

	// print some nice, non-json info about notable repos
	for _, repo := range g.repos {
		ciMsg := "and has NO CI configuration."
		if len(repo.getCISystems()) == 1 {
			ciMsg = fmt.Sprintf("and is configured with %s CI.", repo.getCISystems()[0]) // TODO: include other CI systems
		}
		if len(repo.getCISystems()) > 1 {
			ciMsg = "and is configured with multiple CI systems."
		}
		if len(repo.getTerraformDirs()) > 0 {
			log.Infof("[%s]: contains %d Terraform directories %s\n", repo.Name, len(repo.getTerraformDirs()), ciMsg)
		}
		if len(repo.getCloudformationDirs()) > 0 {
			log.Infof("[%s]: contains %d CloudFormation directories %s\n", repo.Name, len(repo.getCloudformationDirs()), ciMsg)
		}
	}

	var out []Repo
	for _, v := range g.repos {
		out = append(out, v)
	}
	return out, nil
}

// githubCredsFromFS reads the `gh` configuration file on disk to get GitHub credentials.
func githubCredsFromFS() (username string, oauthToken string, retErr error) {
	var ghConfig map[string]interface{}
	var ghConfigFile string

	// Github credentials can exist in one of two files: ~/.config/gh/hosts.yml,
	// or ~/.config/gh/config.yml. Creds _should_ only exist in the latter, but
	// on some systems it appears in the former for whatever reason.
	ghConfigFile, err := homedir.Expand("~/.config/gh/hosts.yml")
	if err != nil {
		retErr = fmt.Errorf("unable to get user Homedir: %w", err)
		return
	}
	if _, err := os.Stat(ghConfigFile); err != nil {
		if !os.IsNotExist(err) {
			retErr = fmt.Errorf("unable to get gh config: %w", err)
			return
		}
		// file does not exist, try another
		ghConfigFile, err = homedir.Expand("~/.config/gh/config.yml")
		if err != nil {
			retErr = fmt.Errorf("unable to get user Homedir: %w", err)
			return
		}
		if _, err := os.Stat(ghConfigFile); err != nil {
			if !os.IsNotExist(err) {
				retErr = fmt.Errorf("unable to get gh config: %w", err)
				return
			}
			// still no file
			retErr = fmt.Errorf("unable to locate GitHub `gh` configuration file")
			return
		}
	}
	yamlf, err := ioutil.ReadFile(ghConfigFile)
	if err != nil {
		retErr = fmt.Errorf("unable to read gh config: %w", err)
		return
	}
	if err := yaml.Unmarshal(yamlf, &ghConfig); err != nil {
		retErr = fmt.Errorf("unable to parse gh config: %w", err)
		return
	}

	n := jnode.FromMap(ghConfig)
	username = n.Path("github.com").Path("user").String()
	oauthToken = n.Path("github.com").Path("oauth_token").String()
	log.Infof(username)
	if username == "" || oauthToken == "" {
		// if empty, we need to descend the hosts tree
		username = n.Path("hosts").Path("github.com").Path("user").String()
		oauthToken = n.Path("hosts").Path("github.com").Path("oauth_token").String()
	}

	if username == "" || oauthToken == "" {
		retErr = fmt.Errorf("unable to find github credentials for scan")
		return
	}
	return
}
