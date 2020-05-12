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

	"github.com/soluble-ai/soluble-cli/pkg/options"
)

type ColumnFormatterType string

var columnFormatters = map[string]options.Formatter{}

func (t ColumnFormatterType) validate() error {
	if _, ok := columnFormatters[string(t)]; !ok {
		return fmt.Errorf("invalid column formatter %s", t)
	}
	return nil
}

func (t ColumnFormatterType) GetFormatter() options.Formatter {
	return columnFormatters[string(t)]
}

func RegisterColumnFormatter(name string, formatter options.Formatter) {
	columnFormatters[name] = formatter
}

func init() {
	RegisterColumnFormatter("", nil)
	RegisterColumnFormatter("ts", options.TimestampFormatter)
	RegisterColumnFormatter("relative_ts", options.RelativeTimestampFormatter)
}
