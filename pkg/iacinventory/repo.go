package iacinventory

// AnalyzedRepos are what we submit to the Soluble API
type AnalyzedRepos struct {
	Repos []Repo `json:"repos"`
}

// Repo is single repository
type Repo interface {
	getName() string
	getCISystems() []string
	getTerraformDirs() []string
	getCloudformationDirs() []string
}
