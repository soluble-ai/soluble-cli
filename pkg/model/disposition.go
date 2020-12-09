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

package model

import (
	"fmt"
)

// What to do with a parameter.  By default the parameter is put
// in the query string (for GET), or in the json body (everything else.)
type ParameterDisposition string

const (
	// Put the parameter value in the context
	ContextDisposition = ParameterDisposition("context")
	// Read a file and post its content as the body
	JSONFileBodyDisposition = ParameterDisposition("json_file_body")
	// Do nothing with the flag
	NOOPDisposition = ParameterDisposition("noop")
)

func (d ParameterDisposition) validate() error {
	if d == "" || d == ContextDisposition || d == JSONFileBodyDisposition || d == NOOPDisposition {
		return nil
	}
	return fmt.Errorf("invalid parameter disposition '%s' must be one of %s, %s",
		d, ContextDisposition, JSONFileBodyDisposition)
}

func (d ParameterDisposition) isDefault() bool {
	return string(d) == ""
}
