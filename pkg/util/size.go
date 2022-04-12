package util

import "fmt"

func Size(n uint64) string {
	switch {
	case n < uint64(1024):
		return fmt.Sprintf("%dB", n)
	case n < 1024*1024:
		return fmt.Sprintf("%dK", n/1024)
	case n < 1024*1024*1024:
		return fmt.Sprintf("%dM", n/(1024*1024))
	default:
		return fmt.Sprintf("%dG", n/(1024*1024*1024))
	}
}
