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

	"github.com/go-resty/resty/v2"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/soluble-ai/soluble-cli/pkg/policy"
	opa "github.com/soluble-ai/soluble-cli/pkg/policy/opa"
	"github.com/soluble-ai/soluble-cli/pkg/provider/terraform"

	"github.com/soluble-ai/soluble-cli/pkg/util"
)

const (
	policyZip = "rego-policies.zip"
	rulePath  = "metadata-opa-policies/policies/opa"
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
	// if the external path is not available use the home space
	if len(policyPath) != 0 {
		policyPath, err = util.GetAbsPath(policyPath)
		if err != nil {
			log.Errorf("unable to get the absolute path for the path %s. Error: %s", policyPath, err.Error())
		}
	} else {
		opts := options.ClientOpts{}
		policyPath = fmt.Sprintf("%s/%s", os.Getenv("HOME"), ".soluble")

		err := os.MkdirAll(policyPath, 0755)
		if err != nil {
			log.Errorf("unable to create a folder for policies with error: %s", err.Error())
		}

		_, err = os.Stat(fmt.Sprintf("%s/%s", policyPath, rulePath))
		if os.IsNotExist(err) {
			apiClient := opts.GetAPIClient()
			// Download the OPA rules from the API server to the specified policyPath
			path := fmt.Sprintf("org/{org}/config/%s", policyZip)

			apiClient.GetClient().SetOutputDirectory(policyPath)

			_, err = apiClient.Get(path, func(req *resty.Request) {
				req.SetOutput(policyZip)
			})
			if err != nil {
				log.Errorf("unable to get the OPA policies, error: %s", err.Error())
			}

			src := fmt.Sprintf("%s/%s", policyPath, policyZip)
			_, err = util.Unzip(src, policyPath)
			if err != nil {
				return nil, err
			}
		}
		policyPath = fmt.Sprintf("%s/%s", policyPath, rulePath)
	}
	log.Infof("Policy Path: %s", policyPath)

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
func (s *Scanner) NewOutputWriter() io.Writer {
	// Color codes will corrupt output, so suppress if not on terminal
	// return termcolor.NewColorizedWriter(os.Stdout)
	return os.Stdout
}
