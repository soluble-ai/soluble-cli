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
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/soluble-ai/go-jnode"
)

var (
	formatterNow       *time.Time
	formatterLocation  *time.Location
	hundredYearsFuture = time.Date(2120, 1, 1, 0, 0, 0, 0, time.UTC)
	timeFormats        = []string{
		time.RFC3339Nano, RFC3339Millis, time.RFC3339, GODefaultFormat,
	}
)

const (
	kb = float64(int64(1) << 10)
	mb = float64(int64(1) << 20)
	gb = float64(int64(1) << 30)

	RFC3339Millis   = "2006-01-02T15:04:05.000Z07:00"
	GODefaultFormat = "2006-01-02 15:04:05.999999999 -0700 MST"
)

type Formatters map[string]Formatter

func (f Formatters) Format(columnName string, n *jnode.Node) string {
	formatter, columnName := f.getFormatter(columnName)
	cell := getCellValue(n, columnName)
	if cell.IsArray() {
		b := strings.Builder{}
		for _, e := range cell.Elements() {
			if b.Len() > 0 {
				b.WriteByte(',')
			}
			b.WriteString(formatter(e))
		}
		return b.String()
	}
	return formatter(cell)
}

func (f Formatters) getFormatter(columnName string) (Formatter, string) {
	if f != nil {
		formatter := f[columnName]
		if formatter != nil {
			return formatter, columnName
		}
	}
	switch {
	case strings.HasSuffix(columnName, "Ts+"):
		cn := columnName[0 : len(columnName)-1]
		return RelativeTimestampFormatter, cn
	case strings.HasSuffix(columnName, "Ts"):
		return TimestampFormatter, columnName
	default:
		return defaultFormatter, columnName
	}
}

func getCellValue(n *jnode.Node, columnName string) *jnode.Node {
	name := bytes.Buffer{}
	wasdot := false
	for _, ch := range columnName {
		switch {
		case ch == '.':
			if wasdot {
				name.WriteRune('.')
				wasdot = false
			} else {
				wasdot = true
			}
		case wasdot:
			n = n.Path(name.String())
			wasdot = false
			name.Reset()
			name.WriteRune(ch)
		default:
			name.WriteRune(ch)
		}
	}
	return n.Path(name.String())
}

func defaultFormatter(n *jnode.Node) string {
	if n.IsNull() {
		return ""
	}
	return n.AsText()
}

func ParseTime(s string) (time.Time, error) {
	for _, f := range timeFormats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("not a time")
}

func TimestampFormatter(n *jnode.Node) string {
	s := n.AsText()
	if formatterLocation == nil {
		formatterLocation = time.Local
	}
	// try and render timestamps in the local timezone
	if t, err := ParseTime(s); err == nil {
		return t.In(formatterLocation).Format(RFC3339Millis)
	}
	return s
}

func RelativeTimestampFormatter(n *jnode.Node) string {
	// render timestamp as relative time
	s := n.AsText()
	if t, err := ParseTime(s); err == nil {
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
		if years > 100 {
			return ">100y"
		}
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

func BytesFormatter(n *jnode.Node) string {
	val := n.AsFloat()
	switch {
	case val >= gb:
		return fmt.Sprintf("%.1fG", val/gb)
	case val >= mb:
		return fmt.Sprintf("%.1fM", val/mb)
	case val >= kb:
		return fmt.Sprintf("%.1fK", val/kb)
	}
	return n.AsText()
}

func NumberFormatter(n *jnode.Node) string {
	if n.GetType() == jnode.Number {
		return fmt.Sprintf("%d", n.AsInt())
	}
	return n.AsText()
}

func DurationMillisFormatter(n *jnode.Node) string {
	if n.GetType() != jnode.Number {
		return n.AsText()
	}
	val := n.AsFloat()
	d := time.Millisecond * time.Duration(val)
	return formatDuration(d)
}

func TruncateFormatter(width int, left bool) Formatter {
	return func(n *jnode.Node) string {
		s := n.AsText()
		if len(s) <= width {
			return s
		}
		if left {
			off := len(s) - width + 3
			return fmt.Sprintf("...%s", s[off:])
		}
		return fmt.Sprintf("%s...", s[0:width-3])
	}
}

func ChopFormatter(width int) Formatter {
	return func(n *jnode.Node) string {
		s := n.AsText()
		if len(s) <= width {
			return s
		}
		return s[0:width]
	}
}
