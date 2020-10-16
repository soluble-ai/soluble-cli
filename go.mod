module github.com/soluble-ai/soluble-cli

go 1.13

require (
	github.com/avast/retry-go v2.6.1+incompatible
	github.com/cyphar/filepath-securejoin v0.2.2
	github.com/fatih/color v1.9.0
	github.com/go-resty/resty/v2 v2.3.0
	github.com/gobwas/glob v0.2.3
	github.com/hashicorp/go-retryablehttp v0.6.6 // indirect
	github.com/hashicorp/go-version v1.2.1
	github.com/hashicorp/hcl/v2 v2.6.0
	github.com/hashicorp/terraform v0.13.4
	github.com/jarcoal/httpmock v1.0.6
	github.com/mattn/go-colorable v0.1.8
	github.com/mitchellh/go-homedir v1.1.0
	github.com/olekukonko/tablewriter v0.0.1
	github.com/open-policy-agent/opa v0.23.2
	github.com/pmezard/go-difflib v1.0.0
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/soluble-ai/go-colorize v0.1.2
	github.com/soluble-ai/go-jnode v0.1.11
	github.com/spf13/afero v1.4.1
	github.com/spf13/cobra v1.0.0
	github.com/zclconf/go-cty v1.5.1
	golang.org/x/net v0.0.0-20200822124328-c89045814202 // indirect
	golang.org/x/sys v0.0.0-20200814200057-3d37ad5750ed // indirect
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/client-go v10.0.0+incompatible
)

replace k8s.io/client-go => k8s.io/client-go v0.19.2
