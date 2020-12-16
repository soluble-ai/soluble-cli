package inventory

import (
	"os"
	"testing"
)

func TestFindGitDir(t *testing.T) {
	if d, _ := FindGitDir("."); d == "" {
		t.Error("can't find git dir")
	}
	if d, _ := FindGitDir(os.TempDir()); d != "" {
		t.Error(d)
	}
}
