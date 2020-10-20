package iacinventory

// Repo is returned by the inventory to be sent to the API as JSON array.
type Repo struct {
	Name               string   `json:"name"`
	CIs                []string `json:"cis"`
	TerraformDirs      []string `json:"terraform_dirs"`
	CloudformationDirs []string `json:"cloudformation_dirs"`
}
