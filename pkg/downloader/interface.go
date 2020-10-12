package downloader

// Downloader helps in downloading different kinds of modules from
// different types of sources
type Downloader interface {
	Download(url, destDir string) (finalDir string, err error)
	DownloadWithType(remoteType, url, dest string) (finalDir string, err error)
	GetURLSubDir(url, dest string) (urlWithType string, subDir string, err error)
	SubDirGlob(string, string) (string, error)
}

// NewDownloader returns a new downloader
func NewDownloader() Downloader {
	return NewGoGetter()
}
