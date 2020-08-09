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
	"fmt"
	"sync"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"github.com/soluble-ai/go-colorize"
)

const (
	Error = iota
	Warning
	Info
	Debug
)

var (
	Level      = Info
	levelNames = map[int]string{
		Error:   "Error",
		Warning: " Warn",
		Info:    " Info",
		Debug:   "Debug",
	}
	lock sync.Mutex
)

func Log(level int, template string, args ...interface{}) {
	if level <= Level {
		lock.Lock()
		defer lock.Unlock()
		colorize.Colorize("{secondary:[%s]} ", levelNames[level])
		colorize.Colorize(template, args...)
		if template[len(template)-1] != '\n' {
			fmt.Fprintln(color.Output)
		}
	}
}

func Infof(template string, args ...interface{}) {
	Log(Info, template, args...)
}

func Debugf(template string, args ...interface{}) {
	Log(Debug, template, args...)
}

func Errorf(template string, args ...interface{}) {
	Log(Error, template, args...)
}

func Warnf(template string, args ...interface{}) {
	Log(Warning, template, args...)
}

func init() {
	color.Output = colorable.NewColorableStderr()
}
