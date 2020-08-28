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

package print

import (
	"fmt"
	"strings"
	"time"

	"github.com/soluble-ai/go-jnode"
)

var (
	formatterNow      *time.Time
	formatterLocation *time.Location

	hundredYearsFuture = time.Date(2120, 1, 1, 0, 0, 0, 0, time.UTC)
)

const (
	kb = float64(int64(1) << 10)
	mb = float64(int64(1) << 20)
	gb = float64(int64(1) << 30)
)

type Formatters map[string]Formatter

func (f Formatters) Format(columnName string, n *jnode.Node) string {
	if f != nil {
		formatter := f[columnName]
		if formatter != nil {
			return formatter(n, columnName)
		}
	}
	var s string
	switch {
	case strings.HasSuffix(columnName, "Ts+"):
		return RelativeTimestampFormatter(n, columnName)
	case strings.HasSuffix(columnName, "Ts"):
		return TimestampFormatter(n, columnName)
	default:
		c := n.Path(columnName)
		if c.IsArray() {
			b := strings.Builder{}
			for _, e := range c.Elements() {
				if b.Len() > 0 {
					b.WriteByte(',')
				}
				b.WriteString(e.AsText())
			}
			s = b.String()
		} else {
			s = c.AsText()
		}
	}
	return s
}

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
		if t.After(hundredYearsFuture) {
			return ">100y"
		}
		d := formatterNow.Sub(t)
		return formatDuration(d)
	}
	return s
}

func formatDuration(d time.Duration) string {
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
	if d < time.Second {
		return fmt.Sprintf("%s%dms", prefix, d.Round(time.Millisecond).Milliseconds())
	}
	return fmt.Sprintf("%s%s", prefix, d.Round(time.Second))
}

func BytesFormatter(n *jnode.Node, columnName string) string {
	val := n.Path(columnName).AsFloat()
	switch {
	case val >= gb:
		return fmt.Sprintf("%.1fG", val/gb)
	case val >= mb:
		return fmt.Sprintf("%.1fM", val/mb)
	case val >= kb:
		return fmt.Sprintf("%.1fK", val/kb)
	}
	return n.Path(columnName).AsText()
}

func NumberFormatter(n *jnode.Node, columnName string) string {
	val := n.Path(columnName).AsFloat()
	return fmt.Sprintf("%d", int64(val))
}

func DurationMillisFormatter(n *jnode.Node, columnName string) string {
	vn := n.Path(columnName)
	if vn.IsMissing() {
		return ""
	}
	val := vn.AsFloat()
	d := time.Millisecond * time.Duration(val)
	return formatDuration(d)
}
