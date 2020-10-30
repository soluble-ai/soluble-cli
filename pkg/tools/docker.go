package tools

import "os/exec"

func HasDocker() error {
	c := exec.Command("docker", "version")
	return c.Run()
}
