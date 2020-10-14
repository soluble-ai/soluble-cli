module github.com/soluble-ai/soluble-cli

go 1.13

require (
	cloud.google.com/go v0.51.0 // indirect
	github.com/accurics/terrascan v1.1.0
	github.com/avast/retry-go v2.6.1+incompatible
	github.com/cyphar/filepath-securejoin v0.2.2
	github.com/fatih/color v1.9.0
	github.com/go-resty/resty/v2 v2.3.0
	github.com/gobwas/glob v0.2.3
	github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1
	github.com/hashicorp/go-getter v1.5.0
	github.com/hashicorp/go-version v1.2.1
	github.com/hashicorp/hcl/v2 v2.6.0
	github.com/hashicorp/terraform v0.13.4
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/jarcoal/httpmock v1.0.6
	github.com/mattn/go-colorable v0.1.8
	github.com/mitchellh/go-homedir v1.1.0
	github.com/olekukonko/tablewriter v0.0.1
	github.com/open-policy-agent/opa v0.23.2
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.0
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749 // indirect
	github.com/shurcooL/vfsgen v0.0.0-20200627165143-92b8a710ab6c // indirect
	github.com/soluble-ai/go-colorize v0.1.2
	github.com/soluble-ai/go-jnode v0.1.11
	github.com/spf13/afero v1.4.1
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.5.1 // indirect
	github.com/zclconf/go-cty v1.5.1
	go.uber.org/zap v1.10.0 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	gopkg.in/resty.v1 v1.12.0
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.19.2 // indirect
	k8s.io/client-go v10.0.0+incompatible
	k8s.io/utils v0.0.0-20200729134348-d5654de09c73 // indirect
)

replace k8s.io/client-go => k8s.io/client-go v0.19.2
