package capture

import (
	"io"
	"os"
	"os/exec"
)

func NewCombinedOutputCapture(stdout, stderr io.Writer) (captureStdout io.Writer, captureStderr io.Writer, cap *Capture) {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	cap = NewCapture()
	captureStdout = io.MultiWriter(stdout, cap)
	captureStderr = io.MultiWriter(stderr, cap)
	return
}

func NewCombinedOutputCaptureForProcess(cmd *exec.Cmd) *Capture {
	stdout, stderr, cap := NewCombinedOutputCapture(cmd.Stdout, cmd.Stderr)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cap
}
