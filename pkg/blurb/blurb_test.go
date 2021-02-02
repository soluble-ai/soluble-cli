package blurb

import "testing"

func TestBlurb(t *testing.T) {
	// not doing much here other than see we don't blow up
	SignupBlurb(nil, "Want to be happy?", "smile more")
}
