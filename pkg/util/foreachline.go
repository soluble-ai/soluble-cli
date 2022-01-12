package util

import (
	"bufio"
	"os"
)

func ForEachLine(path string, fn func(line string) bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if !fn(line) {
			break
		}
	}
	return sc.Err()
}
