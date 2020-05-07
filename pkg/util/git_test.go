package util

import "testing"

func TestGit(t *testing.T) {
	_, err := Git("version")
	if err != nil {
		t.Error(err)
	}
}
