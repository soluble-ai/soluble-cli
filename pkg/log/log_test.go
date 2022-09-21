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

package log

import (
	"bytes"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestLog(t *testing.T) {
	Configure()
	w := bytes.Buffer{}
	color.Output = &w
	color.NoColor = true
	Infof("hello")
	if s := w.String(); s != "[ Info] hello\n" {
		t.Error(s)
	}
}

func TestTemp(t *testing.T) {
	Configure()
	w := bytes.Buffer{}
	color.Output = &w
	color.NoColor = true
	func() {
		t := SetTempLevel(Error)
		defer t.Restore()
		Infof("hello")
	}()
	if s := w.String(); s != "" {
		t.Error(s)
	}
	Infof("There")
	if s := w.String(); s != "[ Info] There\n" {
		t.Error(s)
	}
}

func TestStartupLogging(t *testing.T) {
	assert := assert.New(t)
	configured = false
	w := bytes.Buffer{}
	color.Output = &w
	color.NoColor = true
	Infof("Before")
	assert.Empty(w.String())
	logStartupMessages()
	assert.Equal(w.String(), "[ Info] Before\n")
}
