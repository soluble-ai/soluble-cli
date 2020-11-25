package blurb

import (
	"bytes"
	"os"
	"strings"

	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
)

var Blurbed = false

type authenticatedP interface {
	IsAuthenticated() bool
}

func cap(s string) string {
	return strings.ToUpper(s[0:1]) + s[1:]
}

func SignupBlurb(opts options.Interface, first, use string) {
	if Blurbed {
		return
	}
	_ = os.Stdout.Sync()
	Blurbed = true
	auth := config.Config.APIToken != ""
	if c, ok := opts.(authenticatedP); ok {
		auth = c.IsAuthenticated()
	}
	if use == "" && auth {
		// no need to blurb anything
		return
	}
	s := bytes.Buffer{}
	s.WriteString(first)
	if auth {
		s.WriteString(" ")
		s.WriteString(cap(use))
	} else {
		s.WriteString(" Signup by running {primary:soluble login}")
		if use != "" {
			s.WriteString(" and ")
			s.WriteString(use)
		}
	}
	log.Infof(s.String())
	log.Infof("See {info:%s} for more information", config.Config.GetAppURL())
}
