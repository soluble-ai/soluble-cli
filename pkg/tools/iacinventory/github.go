package iacinventory

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
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

	// get the github.Repositories for the current user
	repos, err := getRepos(g.User, g.OauthToken)
	if err != nil {
		return fmt.Errorf("error getting repositories: %w", err)
	}

	// and map that to our GithubRepo, which implements the Repo interface
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
	for i, repo := range githubRepos {
		log.Infof("[%.3d/%.3d] | [%s]: analyzing repo...\n", i, len(githubRepos), repo.FullName)
		if err := repo.getCode(g.User, g.OauthToken); err != nil {
			log.Infof("[%.3d/%.3d] | [%s]: error fetching archive: %v\n", i, len(githubRepos), repo.FullName, err)
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
			if ci != "" {
				repo.CISystems = append(repo.CISystems, ci)
			}
			if err != nil {
				return err
			}
			tfDir, err := walkTerraformDirs(path, info, err)
			if len(tfDir) != 0 {
				repo.TerraformDirs = append(repo.TerraformDirs, tfDir)
			}
			if err != nil {
				return err
			}
			cfDir, err := walkCloudFormationDirs(path, info, err)
			if err != nil {
				return err
			}
			if len(cfDir) != 0 {
				repo.CloudformationDirs = append(repo.CloudformationDirs, cfDir)
			}
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

func (g *GithubInventorier) Run() ([]GithubRepo, error) {
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

	var out []GithubRepo
	for _, v := range g.repos {
		out = append(out, v)
	}
	return out, nil
}

// githubCredsFromFS reads the `gh` configuration file on disk to get GitHub credentials.
func githubCredsFromFS() (username string, oauthToken string, retErr error) {
	type ghConfigurationData struct {
		ConfigYML struct {
			User  string `yaml:"user,omitempty"`
			Token string `yaml:"oauth_token,omitempty"`
		} `yaml:"github.com,omitempty"`
		HostsYML struct {
			ConfigYML struct {
				User  string `yaml:"user,omitempty"`
				Token string `yaml:"oauth_token,omitempty"`
			} `yaml:"github.com,omitempty"`
		} `yaml:"hosts,omitempty"`
	}

	// Github credentials can exist in one of two files: ~/.config/gh/hosts.yml,
	// or ~/.config/gh/config.yml. Creds _should_ only exist in the latter, but
	// on some systems it appears in the former for whatever reason.

	configFiles := []string{"~/.config/gh/config.yml", "~/.config/gh/hosts.yml"}
	for _, configFile := range configFiles {
		if username != "" && oauthToken != "" {
			break
		}
		ghConfigFile, err := homedir.Expand(configFile)
		if err != nil {
			retErr = fmt.Errorf("unable to get user Homedir: %w", err)
			return
		}
		if _, err := os.Stat(ghConfigFile); err != nil {
			if !os.IsNotExist(err) {
				retErr = fmt.Errorf("unable to get gh config: %w", err)
				return
			}
		}
		yamlf, err := ioutil.ReadFile(ghConfigFile)
		if err != nil {
			retErr = fmt.Errorf("unable to read gh config: %w", err)
			return
		}
		var ghConfig ghConfigurationData
		if err := yaml.Unmarshal(yamlf, &ghConfig); err != nil {
			retErr = fmt.Errorf("unable to parse gh config: %w", err)
			return
		}
		username = ghConfig.ConfigYML.User
		oauthToken = ghConfig.ConfigYML.Token
		if username == "" || oauthToken == "" {
			// if empty, we need to descend the hosts tree
			username = ghConfig.HostsYML.ConfigYML.User
			oauthToken = ghConfig.HostsYML.ConfigYML.Token
		}
	}
	if username == "" || oauthToken == "" {
		retErr = fmt.Errorf("unable to find github credentials for scan")
		return
	}
	return
}
