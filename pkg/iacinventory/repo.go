package iacinventory

// Repo is returned by the inventory to be sent to the API as JSON array.
type Repo struct {
	Name               string   `json:"name"`
	CISystems          []string `json:"ci_systems"`
	TerraformDirs      []string `json:"terraform_dirs"`
	CloudformationDirs []string `json:"cloudformation_dirs"`
}
