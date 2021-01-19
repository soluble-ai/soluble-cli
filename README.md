# Soluble CLI

This is the command line interface for [Soluble](https://soluble.ai).

## Installation

On MacOS use [homebrew](https://brew.sh):

    brew install soluble-ai/soluble/soluble-cli

To upgrade to the latest version:

    brew upgrade soluble-ai/soluble/soluble-cli

On linux, run:

    wget -O - https://raw.githubusercontent.com/soluble-ai/soluble-cli/master/linux-install.sh | sh
    # or
    curl https://raw.githubusercontent.com/soluble-ai/soluble-cli/master/linux-install.sh | sh

The install will drop the executable in the current directory.  If you run this as `root` or can sudo to root,
the install will try to move the binary to `/usr/local/bin/soluble`.

Windows executables can be found on the releases page.

## Run Security Scans

Run security scans on your code with:

    soluble auto-scan -d ~/my-stuff

This will search under `~/my-stuff` for a variety of infrastructre-as-code files and scan them.

If you'd like to manage the findings of those tools with [Soluble](https://app.soluble.cloud), you'll have to authenticate the CLI with:

    soluble login

Then re-run the scan with:

    soluble auto-scan -d ~/my-stuff --upload

You can instead run individual scans:

* `soluble terraform-scan` scans [terraform files](https://www.terraform.io/) with [checkov](https://github.com/bridgecrewio/checkov).  Alternately, `soluble terraform-scan terrascan` or `soluble terraform-scan tfsec` can be used to scan with [terrascan](https://github.com/accurics/terrascan) or [tfsec](https://github.com/tfsec/tfsec) respectively.

* `soluble kubernetes-scan` scans [kubernetes manifest files](https://kubernetes.io/).

* `soluble cloudformation-scan` scans [AWS cloudformation templates](https://aws.amazon.com/cloudformation/) with [cfn-python-lint](https://github.com/aws-cloudformation/cfn-python-lint).  [cfn_nag](https://github.com/stelligent/cfn_nag) and [checkov](https://github.com/bridgecrewio/checkov) are also available.

* `soluble secrets-scan` scans for secrets in code with [detect-secrets](https://github.com/Yelp/detect-secrets).

* `soluble dep-scan` scans application dependencies with [trivy](https://github.com/aquasecurity/trivy).

* `soluble image-scan` scans container images with [trivy](https://github.com/aquasecurity/trivy).

* `soluble code-scan` scans application code with [semgrep](https://semgrep.dev/).

## CI Integration

(WIP - basic notes)

1. Set the environment variable `SOLUBLE_API_TOKEN` to an API token from https://app.soluble.cloud/admin/tokens/access 
2. Add a step to run `soluble auto-scan -d . --upload`, or run individual scans (see `soluble iac-scan --help`)
3. At then end of your CI job, run `soluble build update-pr` to update the pull request with the scan results.

