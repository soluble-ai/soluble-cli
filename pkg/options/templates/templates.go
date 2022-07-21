package templates

import "embed"

//go:embed *.tmpl
var fs embed.FS

func GetEmbeddedTemplate(name string) string {
	dat, err := fs.ReadFile(name)
	if err != nil {
		panic(err)
	}
	return string(dat)
}
