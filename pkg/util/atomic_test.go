package util

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAtomicFileWriter(t *testing.T) {
	assert := assert.New(t)
	dir := t.TempDir()
	a, err := NewAtomicFileWriter(filepath.Join(dir, "atomic_test.txt"))
	assert.NoError(err)
	assert.NoError(a.Close())
	assert.Error(a.Rename())
	ents, err := os.ReadDir(dir)
	assert.NoError(err)
	assert.ElementsMatch(ents, []os.DirEntry{})
}

func TestAtomicFileWriterRename(t *testing.T) {
	assert := assert.New(t)
	dir := t.TempDir()
	a, err := NewAtomicFileWriter(filepath.Join(dir, "atomic_test.txt"))
	assert.NoError(err)
	fmt.Fprintf(a, "hello, world\n")
	assert.NoError(a.Rename())
	assert.NoError(a.Close())
	ents, err := os.ReadDir(dir)
	assert.NoError(err)
	if assert.Equal(1, len(ents)) {
		assert.Equal("atomic_test.txt", ents[0].Name())
	}
	dat, err := os.ReadFile(filepath.Join(dir, "atomic_test.txt"))
	assert.NoError(err)
	assert.Equal("hello, world\n", string(dat))
	b, err := NewAtomicFileWriter(filepath.Join(dir, "atomic_test.txt"))
	assert.NoError(err)
	fmt.Fprintf(b, "howdy, world\n")
	assert.NoError(b.Rename())
	dat, err = os.ReadFile(filepath.Join(dir, "atomic_test.txt"))
	assert.NoError(err)
	assert.Equal("howdy, world\n", string(dat))
}
