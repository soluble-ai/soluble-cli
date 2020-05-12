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

package options

import (
	"fmt"
	"time"

	"github.com/soluble-ai/go-jnode"
)

var (
	formatterNow      *time.Time
	formatterLocation *time.Location
)

func TimestampFormatter(n *jnode.Node, columnName string) string {
	s := n.Path(columnName).AsText()
	if formatterLocation == nil {
		formatterLocation = time.Local
	}
	// try and render timestamps in the local timezone
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.In(formatterLocation).Format(time.RFC3339)
	}
	return s
}

func RelativeTimestampFormatter(n *jnode.Node, columnName string) string {
	// render timestamp as relative time
	if columnName[len(columnName)-1] == '+' {
		columnName = columnName[:len(columnName)-1]
	}
	s := n.Path(columnName).AsText()
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		if formatterNow == nil {
			n := time.Now()
			formatterNow = &n
		}
		d := formatterNow.Sub(t)
		var prefix string
		if d < 0 {
			prefix = "+"
			d = -d
		}
		if d > 365*24*time.Hour {
			years := int64(d) / int64(365*24*time.Hour)
			d -= time.Duration(365*24*years) * time.Hour
			prefix += fmt.Sprintf("%dy", years)
		}
		if d > 24*time.Hour {
			days := int64(d) / int64(24*time.Hour)
			d -= time.Duration(24*days) * time.Hour
			prefix += fmt.Sprintf("%dd", days)
		}
		return prefix + d.Round(time.Second).String()
	}
	return s
}
