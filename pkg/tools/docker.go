package tools

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/log"
)

type DockerTool struct {
	Name            string
	Image           string
	Directory       string
	DockerArgs      []string
	Args            []string
	PolicyDirectory string
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

func (t *DockerTool) run(skipPull bool) ([]byte, error) {
	if err := hasDocker(); err != nil {
		return nil, err
	}
	if !skipPull {
		log.Infof("Pulling {primary:%s}", t.Image)
		// #nosec G204
		pull := exec.Command("docker", "pull", t.Image)
		pull.Stderr = os.Stderr
		pull.Stdout = os.Stdout
		if err := pull.Run(); err != nil {
			log.Warnf("docker pull {primary:%s} failed: {warning:%s}", t.Image, err)
		}
	}
	args := []string{"run", "--rm"}
	if t.Directory != "" {
		args = append(args, "-v", fmt.Sprintf("%s:/src", t.Directory),
			"-w", "/src")
	}
	if t.PolicyDirectory != "" {
		// mount the policy directory and rewrite args
		args = append(args, "-v", fmt.Sprintf("%s:/policy", t.PolicyDirectory))
		for i := range t.Args {
			if t.Args[i] == t.PolicyDirectory {
				t.Args[i] = "/policy"
			}
		}
	}
	args = append(args, t.DockerArgs...)
	args = append(args, t.Image)
	args = append(args, t.Args...)
	run := exec.Command("docker", args...)
	log.Infof("Running {primary:%s}", strings.Join(run.Args, " "))
	run.Stderr = os.Stderr
	return run.Output()
}
