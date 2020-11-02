package download

import "net/http"

type downloadOption func(req *http.Request) error

func withBearerToken(token string) downloadOption {
	return func(req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	}
}
