package tools

import (
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
	c := exec.Command("docker", "version")
	return c.Run()
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
