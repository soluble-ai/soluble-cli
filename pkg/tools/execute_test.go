package tools

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	assert := assert.New(t)
	cmd := exec.Command("ls", "-l", ".", "does-not-exist")
	res := executeCommand(cmd)
	assert.NotNil(res)
	assert.NotEqual(0, res.ExitCode)
	res.ExpectExitCode(0)
	assert.Equal(ExitCodeFailure, res.FailureType)
	assert.NotEmpty(res.FailureMessage)
	assert.NotEmpty(res.Output)
	assert.NotEmpty(res.CombinedOutput)
	assert.Contains(string(res.Output), "execute_test.go")
	cmd = exec.Command("false")
	res = executeCommand(cmd)
	assert.True(res.ExpectExitCode(1))
	assert.Equal(NoFailure, res.FailureType)
}

func TestExecuteParseJSON(t *testing.T) {
	assert := assert.New(t)
	cmd := exec.Command("echo", `{ "hello": "world" }`)
	res := executeCommand(cmd)
	assert.NotNil(res)
	res.ExpectExitCode(0)
	assert.Equal(NoFailure, res.FailureType)
	assert.Equal("", res.FailureMessage)
	assert.True(res.FailureMessage == "")
	n, ok := res.ParseJSON()
	assert.True(ok)
	assert.NotNil(n)
	assert.Equal("world", n.Path("hello").AsText())
}

func TestExecuteError(t *testing.T) {
	assert := assert.New(t)
	failureMessage := "CycleError in aws CycleError: variables form a cycle: aws_instance.nat.provisioner[0].inline"
	cmd := exec.Command("echo",
		`{"level":"fatal","msg":"`+failureMessage+`"}`)
	res := executeCommand(cmd)
	assert.NotNil(res)
	assert.Equal(ExecutionFailure, res.FailureType)
	assert.Equal(failureMessage, res.FailureMessage)
	assert.False(res.ExpectExitCode(0))
}
