package download

import "net/http"

type DownloadOption func(req *http.Request) error

func WithBearerToken(token string) DownloadOption {
	return func(req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	}
}
