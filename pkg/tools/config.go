package tools

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	ignore "github.com/sabhiram/go-gitignore"
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"gopkg.in/yaml.v3"
)

type Config struct {
	path   string
	data   *jnode.Node
	ignore *ignore.GitIgnore
}

func (c *Config) IsIgnored(path string) bool {
	if c.ignore == nil {
		v := c.data.Path("ignore")
		if v.IsArray() {
			lines := make([]string, v.Size())
			for i, line := range v.Elements() {
				lines[i] = line.AsText()
			}
			c.ignore = ignore.CompileIgnoreLines(lines...)
		} else {
			c.ignore = ignore.CompileIgnoreLines(v.AsText())
		}
		if c.ignore == nil {
			log.Warnf("{warning:Invalid ignore lines} {secondary:in %s}", c.path)
		}
	}
	return c.ignore.MatchesPath(path)
}

func ReadConfig(dir string) *Config {
	c := &Config{}
	path := filepath.Join(dir, "config.yml")
	d, err := ioutil.ReadFile(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Warnf("Could not read {warning:%s} - {warning:%s}", path, err)
		}
		return c
	}
	var m map[string]interface{}
	err = yaml.Unmarshal(d, &m)
	if err != nil {
		log.Warnf("Could not parsed {warning:%s} - {warning:%s}", path, err)
	}
	c.data = jnode.FromMap(m)
	c.path = path
	return c
}
