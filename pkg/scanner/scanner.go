// Copyright 2020 Soluble Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scanner

import (
	"fmt"
	"io"
	"os"

	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/policy"
	opa "github.com/soluble-ai/soluble-cli/pkg/policy/opa"
	"github.com/soluble-ai/soluble-cli/pkg/provider/terraform"
)

// Scanner object
type Scanner struct {
	filePath     string
	dirPath      string
	policyPath   string
	terraform    terraform.Terraform
	policyEngine policy.Engine
}

// NewScanner creates a runtime object
func NewScanner(filePath, dirPath, policyPath string) (s *Scanner, err error) {
	fmt.Println(dirPath)
	s = &Scanner{
		filePath:   filePath,
		dirPath:    dirPath,
		policyPath: policyPath,
	}

	// initialize executor
	if err = s.Init(); err != nil {
		return s, err
	}

	return s, nil
}

// Init validates input and initializes iac and cloud providers
func (s *Scanner) Init() error {
	//TODO: validate inputs
	// err := s.ValidateInputs()
	// if err != nil {
	// 	return err
	// }

	// create a new policy engine based on IaC type
	var err error
	s.policyEngine, err = opa.NewEngine(s.policyPath)
	if err != nil {
		log.Errorf("failed to create policy engine. error: '%s'", err)
		return err
	}

	log.Debugf("initialized scanner")
	return nil
}

// Execute validates the inputs, processes the IaC, creates json output
func (s *Scanner) Execute() (results Output, err error) {
	if s.filePath != "" {
		results.ResourceConfig, err = s.terraform.LoadIacFile(s.filePath)
	} else {
		results.ResourceConfig, err = s.terraform.LoadIacDir(s.dirPath)
	}
	if err != nil {
		return results, err
	}
	// evaluate policies
	results.Violations, err = s.policyEngine.Evaluate(policy.EngineInput{InputData: &results.ResourceConfig})
	if err != nil {
		return results, err
	}
	// successful
	return results, nil
}

// NewOutputWriter gets a new io.Writer based on os.Stdout.
// If param useColors=true, the writer will colorize the output
func (s *Scanner) NewOutputWriter() io.Writer {
	// Color codes will corrupt output, so suppress if not on terminal
	// return termcolor.NewColorizedWriter(os.Stdout)
	return os.Stdout
}
