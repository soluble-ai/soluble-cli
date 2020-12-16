package login

import (
	"fmt"
	"os"
)

type HeadlessLeg struct{}

var _ AuthCodeLeg = &HeadlessLeg{}

func (h *HeadlessLeg) GetCode(appURL, state string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/auth/cli-login/%s/0", appURL, state)
	fmt.Printf(`To complete the login process, open the following URL in your browser:

  %s

Paste the resulting authorization code here: `, url)
	_ = os.Stdout.Sync()
	var code string
	_, err := fmt.Scanln(&code)
	return code, err
}
