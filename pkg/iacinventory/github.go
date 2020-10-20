package iacinventory

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"gopkg.in/yaml.v2"
)

var _ IacInventorier = &GithubScanner{}

type GithubScanner struct {
	User       string `yaml:"user"`
	OauthToken string `yaml:"oauth_token"`

	repos map[string]GithubRepo
}

// GetRepos fetches (and expands) the repositories associated with a github account
func (g *GithubScanner) GetRepos() error {
	if g.User == "" || g.OauthToken == "" {
		return fmt.Errorf("no credentials provided")
	}

	ctx := context.TODO()
	client := &http.Client{}
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/repos?per_page=100", nil) // TODO: implement pagination
	req.Header.Set("Authorization", "token "+g.OauthToken)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending requsting during inventory")
	}

	var githubRepos []GithubRepo
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&githubRepos); err != nil {
		resp.Body.Close()
		return fmt.Errorf("error decoding GitHub response to JSON during inventory")
	}
	resp.Body.Close()

	// reposWithCode are the fully initialized repos, including source code
	reposWithCode := make(map[string]GithubRepo)
	for _, repo := range githubRepos {
		// TODO: handle print output using tables and the like
		log.Infof("[%s]: analyzing repo...\n", repo.FullName)
		if err := repo.GetMaster(ctx, client, g.User, g.OauthToken); err != nil {
			// TODO: handle print output using tables and the like
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
			ci, err := WalkCI(path, info, err)
			if err != nil {
				return err
			}
			if len(ci) != 0 {
				repo.CI = append(repo.CI, ci)
			}
			if err := repo.GetTerraformDirs(path, info, err); err != nil {
				return err
			}
			if err := repo.GetCloudFormationDirs(path, info, err); err != nil {
				return err
			}
			return nil
		}); err != nil {
			// TODO: handle print output using tables and the like
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

func (g *GithubScanner) Run() (*jnode.Node, error) {
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
	if err := g.GetRepos(); err != nil {
		return nil, fmt.Errorf("error fetching github repositories: %w", err)
	}

	// massage the data into a sensible format for submission to the API.
	out := make([]Repo, len(g.repos))
	for name, repo := range g.repos {
		var tfdirs []string
		for dir := range repo.TerraformDirs {
			tfdirs = append(tfdirs, dir)
		}
		var cfdirs []string
		for dir := range repo.CloudformationDirs {
			cfdirs = append(cfdirs, dir)
		}
		var cis []string
		for _, ci := range repo.CI {
			cis = append(cis, string(ci))
		}
		out = append(out, Repo{
			Name:               name,
			CIs:                cis,
			TerraformDirs:      tfdirs,
			CloudformationDirs: cfdirs,
		})
	}

	pretty, err := json.MarshalIndent(out, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	// TODO: make outputs selectable by -o flag
	fmt.Println(string(pretty))

	// print some nice, non-json info about notable repos
	// TODO: make selectable by -o flag
	for _, repo := range out {
		ciMsg := "and has NO CI configuration."
		if len(repo.CIs) == 1 {
			ciMsg = fmt.Sprintf("and is configured with %s CI.", repo.CIs[0]) // TODO: include other CI systems
		}
		if len(repo.CIs) > 1 {
			ciMsg = "and is configured with multiple CI systems."
		}
		if len(repo.TerraformDirs) > 0 {
			fmt.Printf("[%s]: contains %d Terraform directories %s\n", repo.Name, len(repo.TerraformDirs), ciMsg)
		}
		if len(repo.CloudformationDirs) > 0 {
			fmt.Printf("[%s]: contains %d CloudFormation directories %s\n", repo.Name, len(repo.CloudformationDirs), ciMsg)
		}
	}
	return nil, nil
}

func (g *GithubScanner) Stop() error {
	return nil
}

// githubCredsFromFS reads the `gh` configuration file on disk to get GitHub credentials.
func githubCredsFromFS() (username string, oauthToken string, retErr error) {
	var ghConfig map[string]interface{}
	var ghConfigFile string

	user, err := user.Current()
	if err != nil {
		retErr = fmt.Errorf("unable to get OS user: %w", err)
		return
	}

	// Github credentials can exist in one of two files: ~/.config/gh/hosts.yml,
	// or ~/.config/gh/config.yml. Creds _should_ only exist in the latter, but
	// on some systems it appears in the former for whatever reason.
	ghConfigFile = filepath.Join(user.HomeDir, ".config/gh/hosts.yml")
	if _, err := os.Stat(filepath.Join(user.HomeDir, ".config/gh/hosts.yml")); err != nil {
		if !os.IsNotExist(err) {
			retErr = fmt.Errorf("unable to get gh config: %w", err)
			return
		}
		// file does not exist, try another
		ghConfigFile = filepath.Join(user.HomeDir, ".config/gh/config.yml")
		if _, err := os.Stat(filepath.Join(user.HomeDir, ".config/gh/config.yml")); err != nil {
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

	// Depending on the "type" of gh configuration (see file selector above),
	// the yaml will be structured differently
	if conf := ghConfig["github.com"]; conf != nil {
		// hub format configuration
		c, ok := conf.(map[interface{}]interface{})
		if !ok {
			retErr = fmt.Errorf("error parsing github config")
			return
		}
		username = c["user"].(string)
		oauthToken = c["oauth_token"].(string)
	}
	if conf := ghConfig["hosts"]; conf != nil {
		// gh format configuration
		gconf, ok := conf.(map[interface{}]interface{})
		if !ok {
			retErr = fmt.Errorf("error parsing github config")
			return
		}
		if nestedconf := gconf["github.com"]; conf != nil {
			c, ok := nestedconf.(map[interface{}]interface{})
			if !ok {
				retErr = fmt.Errorf("error parsing github config")
				return
			}
			username = c["user"].(string)
			oauthToken = c["oauth_token"].(string)
		}
	}

	if username == "" || oauthToken == "" {
		retErr = fmt.Errorf("unable to find github credentials for scan")
		return
	}
	return
}
