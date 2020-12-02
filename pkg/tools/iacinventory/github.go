package iacinventory

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/mitchellh/go-homedir"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/print"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type Github struct {
	tools.ToolOpts
	User                 string
	OauthToken           string
	AllRepos             bool
	PublicRepos          bool
	ExplicitRepositories []string
	Orgs                 []string
}

var _ tools.Interface = &Github{}

func (g *Github) Name() string {
	return "github-iac-inventory"
}

func (g *Github) Register(c *cobra.Command) {
	g.ToolOpts.Register(c)
	flags := c.Flags()
	flags.StringVar(&g.User, "gh-username", "", "Github Username")
	flags.StringVar(&g.OauthToken, "gh-oauthtoken", "", "Github OAuthToken")
	flags.BoolVar(&g.AllRepos, "all", false, "Inventory all accessible public and private repositories.")
	flags.BoolVar(&g.PublicRepos, "public", false, "Inventory accessible public repositories.")
	flags.StringSliceVar(&g.ExplicitRepositories, "repository", nil, "Inventory this repository. May be repeated.")
	flags.StringSliceVar(&g.Orgs, "org", nil, "Inventory repositories for a specific Organization. May be repeated.")
}

func (g *Github) Run() (*tools.Result, error) {
	if !g.AllRepos && !g.PublicRepos && len(g.ExplicitRepositories) == 0 {
		return nil, fmt.Errorf("no repositories to scan, use --public, --all, or --repository")
	}
	if (g.AllRepos || g.PublicRepos) && len(g.ExplicitRepositories) > 0 {
		return nil, fmt.Errorf("use either --all/--public or --repository not both")
	}

	var err error
	// if no authentication options have been set,
	// we default to reading from the filesystem.
	if g.User == "" && g.OauthToken == "" {
		g.User, g.OauthToken, err = githubCredsFromFS()
		if err != nil {
			return nil, err
		}
	}

	// fetch and analyze repositories
	repos, err := g.scanRepos()
	if err != nil {
		return nil, fmt.Errorf("error fetching github repositories: %w", err)
	}
	n, _ := print.ToResult(repos)
	result := &tools.Result{
		Data:   n,
		Values: map[string]string{"USER": g.User},
	}
	return result, nil
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

// getRepos fetches (and expands) the repositories associated with a github account
func (g *Github) scanRepos() ([]*GithubRepo, error) {
	if g.User == "" || g.OauthToken == "" {
		return nil, fmt.Errorf("no credentials provided")
	}
	repos, err := getRepos(g.OauthToken, g.AllRepos, g.ExplicitRepositories)
	if err != nil {
		return nil, fmt.Errorf("error getting repositories: %w", err)
	}

	// filter out non-org repos (keeps count consistent in the loop below)
	var filteredRepos []*github.Repository
	if len(g.Orgs) == 0 {
		filteredRepos = repos
	} else {
		for _, repo := range repos {
			for _, org := range g.Orgs {
				if org != "" {
					// repo.GetOwner().GetName() does not behave.
					owner := strings.Split(repo.GetFullName(), "/")[0]
					if owner != org {
						// skip repositories that do not match the selected org
						continue
					}
					filteredRepos = append(filteredRepos, repo)
				}
			}
		}
	}

	result := []*GithubRepo{}
	for i, repo := range filteredRepos {
		log.Infof("Analyzing repo {primary:%s} (%d of %d)", repo.GetFullName(), i+1, len(filteredRepos))
		r := &GithubRepo{
			Name:     repo.GetName(),
			FullName: repo.GetFullName(),
			GitRepo:  "github.com/" + repo.GetFullName(),
		}
		if err := r.downloadAndScan(g.OauthToken, repo); err != nil {
			log.Warnf("Failed to scan {warning:%s} - {danger:%s}", repo.GetFullName(), err.Error())
			continue
		}
		result = append(result, r)
	}
	return result, nil
}
