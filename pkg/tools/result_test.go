package tools

import (
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/archive"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestFilesTarball(t *testing.T) {
	dir, _ := ioutil.TempDir("", "toolopt*")
	defer os.RemoveAll(dir)
	createFile(dir, "zero.txt", "zero\n")
	createFile(dir, filepath.FromSlash("one/one.txt"), "one\n")
	result := &Result{
		Directory: dir,
	}
	result.AddFile("zero.txt")
	result.AddFile(filepath.Join(dir, filepath.FromSlash("one/one.txt")))
	f, err := result.createTarball()
	util.Must(err)
	mfs := afero.NewMemMapFs()
	util.Must(archive.UntarReader(f, true, mfs, nil))
	if readFile(mfs, "zero.txt") != "zero\n" {
		t.Error("zero.txt")
	}
	if readFile(mfs, "one/one.txt") != "one\n" {
		t.Error("one.txt")
	}
}

func readFile(fs afero.Fs, path string) string {
	f, err := fs.Open(filepath.FromSlash(path))
	util.Must(err)
	defer f.Close()
	dat, err := ioutil.ReadAll(f)
	util.Must(err)
	return string(dat)
}

func createFile(dir, path, content string) {
	p := filepath.Join(dir, path)
	util.Must(os.MkdirAll(filepath.Dir(p), os.ModePerm))
	f, err := os.Create(p)
	util.Must(err)
	util.Must(util.PropagateCloseError(f, func() error {
		_, err := f.WriteString(content)
		return err
	}))
}

func TestUpload(t *testing.T) {
	assert := assert.New(t)
	tempdir, _ := ioutil.TempDir("", "toolopt*")
	defer os.RemoveAll(tempdir)
	createFile(tempdir, filepath.FromSlash(".soluble/config.yml"), "# config.yml\n")
	createFile(tempdir, filepath.FromSlash(".git/config"), "")
	createFile(tempdir, filepath.FromSlash(".github/CODEOWNERS"), "#\n")
	createFile(tempdir, "README", "hello world\n")
	result := &Result{
		Data:      jnode.NewObjectNode(),
		Directory: tempdir,
		Files:     util.NewStringSetWithValues([]string{"README"}),
	}
	result.AddValue("FOO", "hello")
	opts := &ToolOpts{}
	opts.APIServer = "https://api.example.com"
	opts.APIToken = "xxx"
	opts.Organization = "9999"
	httpmock.ActivateNonDefault(opts.GetAPIClient().GetClient().GetClient())
	httpmock.RegisterResponder("POST", "https://api.example.com/api/v1/xcp/test/data",
		func(h *http.Request) (*http.Response, error) {
			n := jnode.NewObjectNode()
			a := n.PutObject("assessment")
			a.Put("appUrl", "http://app.example.com/A1")
			resp, err := httpmock.NewJsonResponse(http.StatusOK, n)
			assert.Nil(h.ParseMultipartForm(1 << 20))
			checkFile(assert, h, "CODEOWNERS", nil)
			checkFile(assert, h, "tarball", nil)
			checkFile(assert, h, "config.yml", nil)
			assert.Equal(h.FormValue("FOO"), "hello")
			return resp, err
		})
	assert.Nil(result.report(opts, "test"))
	assert.Equal("http://app.example.com/A1", result.AssessmentURL)
}

func checkFile(assert *assert.Assertions, h *http.Request, name string, fn func(*assert.Assertions, multipart.File)) {
	f, _, err := h.FormFile(name)
	if err != nil {
		assert.Fail(err.Error(), "file %s not found in upload", name)
		return
	}
	defer f.Close()
	if fn != nil {
		fn(assert, f)
	}
}
