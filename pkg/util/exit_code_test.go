// +build darwin linux

package util

import (
	"fmt"
	"os/exec"
	"testing"
)

func TestExitCode(t *testing.T) {
	if ExitCode(nil) != 0 {
		t.Error("nil = 0")
	}
	if ExitCode(fmt.Errorf("error")) != -1 {
		t.Error("err != -1")
	}
	if ee := ExitCode(exec.Command("false").Run()); ee != 1 {
		t.Error(ee)
	}
}
