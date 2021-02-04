package tools

import "github.com/soluble-ai/go-jnode"

func PassFormatter(n *jnode.Node) string {
	if n.AsBool() {
		return "PASS"
	}
	return "FAIL"
}
