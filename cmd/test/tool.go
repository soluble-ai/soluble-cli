package test

import (
	"os"
	"testing"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/inventory"
	"github.com/soluble-ai/soluble-cli/pkg/util"
)

type Tool struct {
	Command
	Fingerprints     *jnode.Node
	fingerprintsFile string
	upload           bool
	dir              string
}

func NewTool(t *testing.T, args ...string) *Tool {
	return &Tool{
		Command: Command{
			T:    t,
			Args: args,
		},
	}
}

func (t *Tool) WithUpload(enable bool) *Tool {
	t.upload = enable
	return t
}

func (t *Tool) WithRepoRootDir() *Tool {
	t.T.Helper()
	dir, err := inventory.FindRepoRoot(".")
	if err != nil {
		t.T.Fatal(err)
	}
	t.dir = dir
	return t
}

func (t *Tool) WithFingerprints() *Tool {
	var err error
	t.fingerprintsFile, err = util.TempFile("fingerprints*")
	util.Must(err)
	t.Args = append(t.Args, "--save-fingerprints", t.fingerprintsFile)
	return t
}

func (t *Tool) Run() error {
	if t.fingerprintsFile != "" {
		defer os.Remove(t.fingerprintsFile)
	}
	if t.upload {
		RequireAPIToken(t.T)
	} else {
		t.Args = append(t.Args, "--upload=false")
	}
	// NB - for now disabling custom policies as it fails with the
	// api token in github
	t.Args = append(t.Args, "--no-color", "--disable-custom-policies")
	if t.dir != "" {
		t.Args = append(t.Args, "-d", t.dir)
	}
	if err := t.Command.Run(); err != nil {
		return err
	}
	if t.fingerprintsFile != "" {
		dat, err := os.ReadFile(t.fingerprintsFile)
		util.Must(err)
		t.Fingerprints, err = jnode.FromJSON(dat)
		util.Must(err)
	}
	return nil
}
