package tools

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/api"
	"github.com/soluble-ai/soluble-cli/pkg/capture"
	"github.com/soluble-ai/soluble-cli/pkg/compress"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/redaction"
	"github.com/soluble-ai/soluble-cli/pkg/xcp"
)

type FailureType string

const (
	InstallFailure       = FailureType("install")
	ExecutionFailure     = FailureType("execution")
	GarbledResultFailure = FailureType("garbled_result")
	ExitCodeFailure      = FailureType("exit_code")
	NoFailure            = FailureType("")
)

type ExecuteResult struct {
	Args           []string
	FailureType    FailureType
	FailureMessage string
	ExitCode       int
	CombinedOutput *strings.Builder
	Output         []byte
}

type ExecuteError struct {
	Level logrus.Level `json:"level"`
	Msg   string       `json:"msg"`
}

func executeCommand(cmd *exec.Cmd) *ExecuteResult {
	result := &ExecuteResult{
		Args: cmd.Args,
	}
	var output *bytes.Buffer
	if cmd.Stdout == nil {
		output = &bytes.Buffer{}
		cmd.Stdout = output
	}
	cap := capture.NewCombinedOutputCaptureForProcess(cmd)
	defer cap.Close()
	err := cmd.Run()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.FailureMessage = err.Error()
			result.FailureType = ExecutionFailure
		}
	}
	out, capErr := cap.OutputBytes()
	if capErr != nil {
		log.Warnf("Could not capture output of {info:%s} - {warning:%s}", cmd.Args[0], capErr)
	}
	s := &strings.Builder{}
	if err := redaction.RedactStream(bytes.NewReader(out), s); err != nil {
		log.Warnf("Could not redact output of {info:%s} - {warning:%s}", cmd.Args[0], err)
	}
	result.CombinedOutput = s
	var execErr ExecuteError
	if err := json.Unmarshal(out, &execErr); err == nil {
		if execErr.Msg != "" {
			result.FailureType = ExecutionFailure
			result.FailureMessage = execErr.Msg
			return result
		}
	}
	if output != nil {
		result.Output = output.Bytes()
	}
	return result
}

func (r *ExecuteResult) SetUploadValues(values map[string]string) {
	values["EXIT_CODE"] = fmt.Sprintf("%d", r.ExitCode)
	if r.FailureType != "" {
		values["FAILURE_TYPE"] = string(r.FailureType)
		values["SUCCESS"] = "false"
	} else {
		values["SUCCESS"] = "true"
	}
	if r.FailureMessage != "" {
		values["FAILURE_MESSAGE"] = r.FailureMessage
	}
	values["COMMAND"] = strings.Join(r.Args, " ")
}

func (r *ExecuteResult) AppendUploadOptions(compressFiles bool, options []api.Option) []api.Option {
	if r.CombinedOutput.Len() > 0 {
		var toolLogOpt api.Option
		src := strings.NewReader(r.CombinedOutput.String())
		if compressFiles {
			gz := compress.NewGZIPPipe(src)
			toolLogOpt = xcp.WithFileFromReader("tool_log", "tool.log.gz", gz)
		} else {
			toolLogOpt = xcp.WithFileFromReader("tool_log", "tool.log", src)
		}
		options = append(options, toolLogOpt)
	}
	return options
}

func (r *ExecuteResult) ToResult(dir string) *Result {
	return &Result{
		Directory:     dir,
		ExecuteResult: r,
	}
}

func (r *ExecuteResult) SetFailureFromError(ft FailureType, err error) {
	r.FailureType = ft
	r.FailureMessage = err.Error()
}

func (r *ExecuteResult) ParseJSON() (*jnode.Node, bool) {
	n, err := jnode.FromJSON(r.Output)
	if err != nil {
		r.SetFailureFromError(GarbledResultFailure, err)
		return nil, false
	}
	return n, true
}

func (r *ExecuteResult) ExpectExitCode(codes ...int) bool {
	if r.FailureType == ExecutionFailure {
		return false
	}
	for _, code := range codes {
		if r.ExitCode == code {
			return true
		}
	}
	r.FailureType = ExitCodeFailure
	r.FailureMessage = fmt.Sprintf("process exited with code %d", r.ExitCode)
	return false
}

func (r *ExecuteResult) ToError() error {
	if r.FailureType != "" {
		return fmt.Errorf("%s", r.FailureMessage)
	}
	return nil
}
