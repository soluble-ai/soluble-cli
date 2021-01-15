package githook

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
)

func installCommand() *cobra.Command {
	var (
		force    bool
		noEditor bool
	)
	opts := options.PrintOpts{}
	c := &cobra.Command{
		Use:   "install",
		Short: "Install git hooks in the current repository's .git/hooks/ directory",
		Long: `Install githook templates from github.com/soluble-ai/githooks

For repos with application source code:
  $ soluble githook install app    # github.com/soluble-ai/githooks/tree/master/app

For repos with IaC files:
  $ soluble githook install iac    # github.com/soluble-ai/githooks/tree/master/iac`,
		Args:      cobra.ExactValidArgs(1),
		ValidArgs: []string{"app", "iac"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !isGitRepository() {
				return fmt.Errorf("current directory is not the root of a git repository")
			}
			mgr := download.NewManager()
			dl, err := mgr.Install(&download.Spec{
				Name: "githooks",
				URL:  "https://github.com/soluble-ai/githooks/archive/master.zip",
			})
			if err != nil {
				return fmt.Errorf("error downloading githooks: %w", err)
			}
			hookDir := filepath.Join(dl.Dir, "githooks-master", strings.Join(args, ""))
			hookDirInfo, err := os.Stat(hookDir)
			if err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("githook directory does not exist: %q", hookDir)
				}
				return fmt.Errorf("error checking githook directory: %w", err)
			}
			if !hookDirInfo.IsDir() {
				return fmt.Errorf("githook directory was file, not directory: %q", hookDir)
			}
			err = filepath.Walk(hookDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				hookDat, err := ioutil.ReadFile(path)
				if err != nil {
					return fmt.Errorf("error reading hook file %q: %w", path, err)
				}
				hookFile := filepath.Join(".git/hooks", info.Name())
				_, err = os.Stat(hookFile)
				if os.IsNotExist(err) || force {
					if err := ioutil.WriteFile(hookFile, hookDat, 0o755); err != nil { //nolint: gosec // executable permissions are expected here
						return fmt.Errorf("error writing hook file: %q: %w", hookFile, err)
					}
					log.Infof("wrote hook file: %q", hookFile)
				} else {
					if err != nil {
						return fmt.Errorf("error stating hook file: %q: %w", path, err)
					}
					return fmt.Errorf("hook files already exist. To force update, re-run with --force")
				}
				return nil
			})
			return err
		},
	}
	opts.Register(c)
	c.Flags().BoolVar(&force, "force", false, "Force installation of hooks, overwriting existing hooks")
	c.Flags().BoolVar(&noEditor, "no-editor", false, "Do not automatically open $EDITOR for installed hooks")
	return c
}

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:     "githook",
		Aliases: []string{"githooks", "hooks"},
		Short:   "Manage git hooks",
	}
	c.AddCommand(installCommand())
	return c
}

func isGitRepository() bool {
	f, err := os.Stat(".git")
	if os.IsNotExist(err) {
		return false
	}
	if f.IsDir() {
		return true
	}
	return false
}
