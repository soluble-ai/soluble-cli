package githook

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/download"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
)

func installCommand() *cobra.Command {
	var (
		force      bool
		editor     bool
		appendFile bool
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
			var installedHooks []string
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
				switch {
				case os.IsNotExist(err) || force:
					if err := ioutil.WriteFile(hookFile, hookDat, 0o755); err != nil { //nolint: gosec // executable permissions are expected here
						return fmt.Errorf("error writing hook file: %q: %w", hookFile, err)
					}
					log.Infof("wrote hook file: %q", hookFile)
					installedHooks = append(installedHooks, hookFile)
				case appendFile:
					log.Infof("appending to hook file: %q", hookFile)
					f, err := os.OpenFile(hookFile,
						os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
					if err != nil {
						return fmt.Errorf("error opening file for append %q: %w", hookFile, err)
					}
					defer f.Close()
					scanner := bufio.NewScanner(bytes.NewBuffer(hookDat))
					scanner.Split(bufio.ScanLines)
					var hookDatNoShebang []byte
					for scanner.Scan() {
						if bytes.HasPrefix(scanner.Bytes(), []byte("#!")) {
							continue
						}
						hookDatNoShebang = append(hookDatNoShebang, scanner.Bytes()...)
						hookDatNoShebang = append(hookDatNoShebang, '\n')
					}
					_, _ = f.WriteString("\n")
					_, err = f.Write(hookDatNoShebang)
					if err != nil {
						return fmt.Errorf("error appending to hook file %q: %w", hookFile, err)
					}
					installedHooks = append(installedHooks, hookFile)
					// we may be opening the files in an editor next
					_ = f.Sync()
				default:
					if err != nil {
						return fmt.Errorf("error stating hook file: %q: %w", path, err)
					}
					return fmt.Errorf("hook files already exist. To append, re-run with --append. To overwrite, re-run with --force")
				}
				if editor {
					v, exists := os.LookupEnv("EDITOR")
					if !exists {
						return fmt.Errorf("the $EDITOR environment variable isn't set")
					}
					cmd := exec.Command(v, installedHooks...)
					cmd.Stdin = os.Stdin
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					return cmd.Run()
				}
				return nil
			})
			return err
		},
	}
	opts.Register(c)
	c.Flags().BoolVar(&force, "force", false, "Force installation of hooks, overwriting existing hooks")
	c.Flags().BoolVar(&editor, "editor", false, "Automatically open $EDITOR for installed hooks")
	c.Flags().BoolVar(&appendFile, "append", false, "Append to file if it exists")
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
