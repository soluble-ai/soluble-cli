package util

import "strings"

// Replaces newlines in s with spaces and truncates with "..."
// if the len(s) is more than m.  The resulting string will
// always be m characters or less.
func TruncateRight(s string, m int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > m {
		return s[0:m-3] + "..."
	}
	return s
}
