module github.com/soluble-ai/soluble-cli

go 1.13

require (
	github.com/accurics/terrascan v1.1.0
	github.com/avast/retry-go v2.6.1+incompatible
	github.com/cyphar/filepath-securejoin v0.2.2
	github.com/fatih/color v1.9.0
	github.com/go-resty/resty/v2 v2.3.0
	github.com/gobwas/glob v0.2.3
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
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749 // indirect
	github.com/shurcooL/vfsgen v0.0.0-20200824052919-0d455de96546 // indirect
	github.com/soluble-ai/go-colorize v0.1.2
	github.com/soluble-ai/go-jnode v0.1.11
	github.com/spf13/afero v1.4.1
	github.com/spf13/cobra v1.0.0
	github.com/zclconf/go-cty v1.5.1
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/client-go v10.0.0+incompatible
)

replace k8s.io/client-go => k8s.io/client-go v0.19.2
