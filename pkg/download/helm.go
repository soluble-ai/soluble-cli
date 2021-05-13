package download

import (
	"fmt"
	"strings"

	"github.com/google/go-github/v32/github"
)

func getHelmDownloadURL(asset *github.ReleaseAsset) string {
	url := asset.GetBrowserDownloadURL()
	slash := strings.LastIndex(url, "/")
	dot := strings.LastIndex(url, ".")
	return fmt.Sprintf("https://get.helm.sh/%s", url[slash+1:dot])
}
