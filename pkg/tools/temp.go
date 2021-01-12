package tools

import (
	"io/ioutil"
	"os"
)

func TempFile(pattern string) (name string, err error) {
	var f *os.File
	f, err = ioutil.TempFile("", pattern)
	if err != nil {
		return
	}
	name = f.Name()
	f.Close()
	return
}
