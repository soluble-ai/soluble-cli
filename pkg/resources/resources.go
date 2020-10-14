package resources

var (
	rootPath string
)

func RootPath() string {
	if rootPath == "" {
		return "<internal>"
	}
	return rootPath
}
