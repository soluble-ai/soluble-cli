package util

import "os/exec"

// Returns the exit code from an error if the error is an exec.ExitError,
// or -1 if the err is non-nil, or 0 otherwise.
func ExitCode(err error) int {
	if ee, ok := err.(*exec.ExitError); ok {
		return ee.ExitCode()
	}
	if err != nil {
		return -1
	}
	return 0
}
