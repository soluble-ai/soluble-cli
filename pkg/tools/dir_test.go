package tools

import (
	"path/filepath"
	"testing"
)

func TestGetDirectory(t *testing.T) {
	o := &DirectoryBasedToolOpts{}
	dir := o.GetDirectory()
	if !filepath.IsAbs(dir) {
		t.Error(dir)
	}
}
