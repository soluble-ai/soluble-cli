package redaction

import (
	"bufio"
	"io"
	"strings"

	"github.com/owenrumney/squealer/pkg/squealer"
)

var (
	scanner  = squealer.NewStringScanner()
	redacted = []byte("  *** redacted ***\n")
	newline  = []byte("\n")
)

func ContainsSecret(s string) bool {
	res := scanner.Scan(s)
	return res.TransgressionFound
}

func RedactStream(r io.Reader, w io.Writer) error {
	sc := bufio.NewScanner(r)
	redactCount := 0
	for sc.Scan() {
		line := sc.Text()
		if ContainsSecret(line) {
			if isBlockSecret(line) {
				redactCount += 5
			} else {
				redactCount += 1
			}
		}
		if redactCount > 0 {
			redactCount -= 1
			if _, err := w.Write(redacted); err != nil {
				return err
			}
		} else {
			if _, err := w.Write([]byte(line)); err != nil {
				return err
			}
			if _, err := w.Write(newline); err != nil {
				return err
			}
		}
	}
	return sc.Err()
}

func isBlockSecret(line string) bool {
	return strings.Contains(line, "-----")
}
