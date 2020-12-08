package semgrep

import (
	"fmt"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/cobra"
)

type Tool struct {
	tools.DirectoryBasedToolOpts
	Pattern string
	Lang    string
	Config  string

	extraArgs []string
}

var _ tools.Interface = &Tool{}

func (*Tool) Name() string {
	return "semgrep"
}

func (t *Tool) Register(cmd *cobra.Command) {
	t.DirectoryBasedToolOpts.Register(cmd)
	flags := cmd.Flags()
	flags.StringVarP(&t.Pattern, "pattern", "e", "", "Code search pattern.")
	flags.StringVarP(&t.Lang, "lang", "l", "", "Parse pattern and all files in specified language. Must be used with -e/--pattern.")
	flags.StringVarP(&t.Config, "config", "f", "", "YAML configuration file, directory of YAML files ending in .yml|.yaml, URL of a configuration file, or semgrep registry entry name.")
}

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "semgrep",
		Short: "Run semgrep",
		Long:  "Run semgrep in a docker container against a directory.  Any additional arguments will be passed onwards.",
		Example: `# get help
... semgrep -- --help

# look for append literal string to a string buffer (silly example)
... semgrep -e '$SB.append("...")' -l java`,
		Args: func(cmd *cobra.Command, args []string) error {
			t.extraArgs = args
			return nil
		},
	}
}

func (t *Tool) Run() (*tools.Result, error) {
	args := []string{"--json"}
	if t.Pattern != "" {
		args = append(args, "-e", t.Pattern)
	}
	if t.Lang != "" {
		args = append(args, "-l", t.Lang)
	}
	if t.Config != "" {
		args = append(args, "-f", t.Config)
	}
	args = append(args, t.extraArgs...)
	args = append(args, ".")
	d, err := t.RunDocker(&tools.DockerTool{
		Image:     "returntocorp/semgrep:latest",
		Directory: t.GetDirectory(),
		Args:      args,
	})
	if err != nil && util.ExitCode(err) != 1 {
		// semgrep exits 1 if it finds issues
		return nil, err
	}
	n, err := jnode.FromJSON(d)
	if err != nil {
		fmt.Println(string(d))
		if util.StringSliceContains(t.extraArgs, "--help") {
			return nil, nil
		}
		return nil, fmt.Errorf("could not parse JSON: %w", err)
	}
	result := &tools.Result{
		Data:      n,
		PrintPath: []string{"results"},
		PrintColumns: []string{
			"check_id", "extra.severity", "extra.message", "path",
		},
	}
	return result, nil
}
