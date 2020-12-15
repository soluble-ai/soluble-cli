package inventory

import (
	"os"
	"testing"
)

func TestFindGitDir(t *testing.T) {
	if d, _ := FindGitDir(); d == "" {
		t.Error("can't find git dir")
	}
	dir, _ := os.Getwd()
	defer func() { _ = os.Chdir(dir) }()
	_ = os.Chdir(os.TempDir())
	if d, _ := FindGitDir(); d != "" {
		t.Error(d)
	}
}
