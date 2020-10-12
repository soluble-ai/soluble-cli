package downloader

import (
	cleanhttp "github.com/hashicorp/go-cleanhttp"
	getter "github.com/hashicorp/go-getter"
)

// GoGetter implements the Downloader interface
type GoGetter struct{}

// list of supported detectors
var goGetterDetectors = []getter.Detector{
	new(getter.GitHubDetector),
	new(getter.GitDetector),
	new(getter.BitBucketDetector),
	new(getter.GCSDetector),
	new(getter.S3Detector),
	new(getter.FileDetector),
}

// empty list of detectors
var goGetterNoDetectors = []getter.Detector{}

// list of supported decompressors
var goGetterDecompressors = map[string]getter.Decompressor{
	"bz2": new(getter.Bzip2Decompressor),
	"gz":  new(getter.GzipDecompressor),
	"xz":  new(getter.XzDecompressor),
	"zip": new(getter.ZipDecompressor),

	"tar.bz2":  new(getter.TarBzip2Decompressor),
	"tar.tbz2": new(getter.TarBzip2Decompressor),

	"tar.gz": new(getter.TarGzipDecompressor),
	"tgz":    new(getter.TarGzipDecompressor),

	"tar.xz": new(getter.TarXzDecompressor),
	"txz":    new(getter.TarXzDecompressor),
}

// list of supported getters
var goGetterGetters = map[string]getter.Getter{
	"file":  new(getter.FileGetter),
	"gcs":   new(getter.GCSGetter),
	"git":   new(getter.GitGetter),
	"hg":    new(getter.HgGetter),
	"s3":    new(getter.S3Getter),
	"http":  getterHTTPGetter,
	"https": getterHTTPGetter,
}

var getterHTTPClient = cleanhttp.DefaultClient()

var getterHTTPGetter = &getter.HttpGetter{
	Client: getterHTTPClient,
	Netrc:  true,
}
