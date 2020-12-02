package tools

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/log"
)

type DockerTool struct {
	Name       string
	Image      string
	DockerArgs []string
	Args       []string
}

func hasDocker() error {
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

func (t *DockerTool) run() ([]byte, error) {
	if err := hasDocker(); err != nil {
		return nil, err
	}
	// #nosec G204
	pull := exec.Command("docker", "pull", t.Image)
	if err := pull.Run(); err != nil {
		log.Warnf("docker pull {primary:%s} failed: {warning:%s}", t.Image, err)
	}
	args := append([]string{"run"}, t.DockerArgs...)
	args = append(args, t.Image)
	args = append(args, t.Args...)
	run := exec.Command("docker", args...)
	log.Infof("Running {primary:%s}", strings.Join(run.Args, " "))
	run.Stderr = os.Stderr
	return run.Output()
}
