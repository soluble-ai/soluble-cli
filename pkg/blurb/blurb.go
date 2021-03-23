// Copyright 2021 Soluble Inc
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

package blurb

import (
	"bytes"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
)

var Blurbed = false

type authenticatedP interface {
	IsAuthenticated() bool
}

func cap(s string) string {
	return strings.ToUpper(s[0:1]) + s[1:]
}

func SignupBlurb(opts options.Interface, first, use string) {
	if Blurbed {
		return
	}
	Blurbed = true
	auth := config.Config.GetAPIToken() != ""
	if c, ok := opts.(authenticatedP); ok {
		auth = c.IsAuthenticated()
	}
	if use == "" && auth {
		// no need to blurb anything
		return
	}
	s := bytes.Buffer{}
	s.WriteString(first)
	if auth {
		s.WriteString(" ")
		s.WriteString(cap(use))
	} else {
		s.WriteString(" Signup by running {primary:soluble login}")
		if use != "" {
			s.WriteString(" and ")
			s.WriteString(use)
		}
	}
	log.Infof(s.String())
	log.Infof("See {info:%s} for more information", config.Config.GetAppURL())
}
