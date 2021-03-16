package inventory

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/util"
)

type LanguageDetector struct {
	getValues          func(*Manifest) *util.StringSet
	markerFiles        []string
	content            ContentDetector
	collapseNestedDirs bool
}

var _ FileDetector = &LanguageDetector{}
var _ FinalizeDetector = &LanguageDetector{}

func (d *LanguageDetector) DetectFileName(m *Manifest, path string) ContentDetector {
	base := filepath.Base(path)
	for _, mf := range d.markerFiles {
		if mf == base {
			d.getValues(m).Add(filepath.Dir(path))
			return d.content
		}
	}
	return nil
}

func (d *LanguageDetector) FinalizeDetection(m *Manifest) {
	if !d.collapseNestedDirs {
		return
	}
	// collapse nested directories with markers into single dir
	values := d.getValues(m)
	if values.Len() == 0 {
		return
	}
	dirs := values.Values()
	values.Reset()
	sort.Strings(dirs)
	values.Add(dirs[0])
	p := dirs[0]
	for i := 1; i < len(dirs); i++ {
		if p != "." && (!strings.HasPrefix(dirs[i], p) || (len(dirs[i]) > len(p) && dirs[i][len(p)] != os.PathSeparator)) {
			values.Add(dirs[i])
			p = dirs[i]
		}
	}
}

func pythonDetector() *LanguageDetector {
	return &LanguageDetector{
		getValues:   func(m *Manifest) *util.StringSet { return &m.PythonDirectories },
		markerFiles: []string{"Pipfile", "requirements.txt"},
	}
}

func goDetector() *LanguageDetector {
	return &LanguageDetector{
		getValues:   func(m *Manifest) *util.StringSet { return &m.GODirectories },
		markerFiles: []string{"go.mod"},
	}
}

func javaAntMavenDetector() *LanguageDetector {
	return &LanguageDetector{
		getValues:          func(m *Manifest) *util.StringSet { return &m.JavaDirectories },
		markerFiles:        []string{"pom.xml", "build.xml", "build.gradle"},
		collapseNestedDirs: true,
	}
}

func javaGradleDetector() *LanguageDetector {
	return &LanguageDetector{
		getValues:   func(m *Manifest) *util.StringSet { return &m.JavaDirectories },
		markerFiles: []string{"build.gradle"},
	}
}

func rubyDetector() *LanguageDetector {
	return &LanguageDetector{
		getValues:   func(m *Manifest) *util.StringSet { return &m.RubyDirectories },
		markerFiles: []string{"Gemfile"},
	}
}

func nodeDetector() *LanguageDetector {
	return &LanguageDetector{
		getValues:   func(m *Manifest) *util.StringSet { return &m.NodeDirectories },
		markerFiles: []string{"package-lock.json", "yarn.lock"},
	}
}
