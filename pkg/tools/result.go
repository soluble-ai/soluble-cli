package tools

import (
	"bytes"
	"io"
	"os"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/archive"
	"github.com/soluble-ai/soluble-cli/pkg/client"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/soluble-ai/soluble-cli/pkg/xcp"
	"github.com/spf13/afero"
)

type Result struct {
	Data         *jnode.Node
	Values       map[string]string
	Directory    string
	Files        *util.StringSet
	PrintPath    []string
	PrintColumns []string

	AssessmentURL string
}

func (r *Result) AddFile(path string) *Result {
	if r.Files == nil {
		r.Files = util.NewStringSet()
	}
	r.Files.Add(path)
	return r
}

func (r *Result) AddValue(name, value string) *Result {
	if r.Values == nil {
		r.Values = map[string]string{}
	}
	r.Values[name] = value
	return r
}

func (r *Result) Report(tool Interface) error {
	rr := bytes.NewReader([]byte(r.Data.String()))
	log.Infof("Uploading results of {primary:%s}", tool.Name())
	options := []client.Option{
		xcp.WithCIEnv, xcp.WithFileFromReader("results_json", "results.json", rr),
	}
	o := tool.GetToolOptions()
	if !o.OmitContext && r.Files != nil {
		tarball, err := r.createTarball()
		if err != nil {
			return err
		}
		defer tarball.Close()
		defer os.Remove(tarball.Name())
		options = append(options, xcp.WithFileFromReader("tarball", "context.tar.gz", tarball))
	}
	n, err := o.GetAPIClient().XCPPost(o.GetOrganization(), tool.Name(), nil, r.Values, options...)
	if err != nil {
		return err
	}
	r.AssessmentURL = n.Path("assessment").Path("appUrl").AsText()
	return nil
}

func (r *Result) createTarball() (afero.File, error) {
	fs := afero.NewOsFs()
	f, err := afero.TempFile(fs, "", "soluble-cli*")
	if err != nil {
		return nil, err
	}
	tar := archive.NewTarballWriter(f)
	err = util.PropagateCloseError(tar, func() error {
		if r.Files != nil {
			for _, file := range r.Files.Values() {
				if err := tar.WriteFile(fs, r.Directory, file); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return nil, err
	}
	// leave the tarball open, but rewind it to the start
	_, err = f.Seek(0, io.SeekStart)
	return f, err
}
