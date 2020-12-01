package tools

import (
	"fmt"
	"os/exec"
)

func HasDocker() error {
	// "The operating-system independent way to check whether Docker is running
	// is to ask Docker, using the docker info command."
	// ref: https://docs.docker.com/config/daemon/#check-whether-docker-is-running
	c := exec.Command("docker", "info")
	err := c.Run()
	switch c.ProcessState.ExitCode() {
	case 0:
		return nil
	case 1:
		return fmt.Errorf("the docker server is not available: %w", err)
	case 127:
		return fmt.Errorf("the docker executable is not present, or is not in the PATH: %w", err)
	default:
		return fmt.Errorf("unknown error checking docker availability: %w", err)
	}
}
