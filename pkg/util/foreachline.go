package util

import (
	"bufio"
	"io"
	"os"
)

func ForEachLine(path string, fn func(line string) bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return ForEachReaderLine(f, fn)
}

func ForEachReaderLine(r io.Reader, fn func(line string) bool) error {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if !fn(line) {
			break
		}
	}
	return sc.Err()
}
