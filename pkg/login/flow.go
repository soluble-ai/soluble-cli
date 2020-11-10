package login

import "strings"

type Response struct {
	UserID    string `json:"userId"`
	OrgID     string `json:"orgId"`
	Token     string `json:"token"`
	APIServer string `json:"apiServer"`
}

type Flow interface {
	Run() (*Response, error)
	Close() error
}

func (r *Response) defaultAPIServer(appURL string) {
	const httpsAPI = "https://app."
	if strings.HasPrefix(appURL, httpsAPI) {
		r.APIServer = "https://api." + appURL[len(httpsAPI):]
	}
}
