package util

import "github.com/soluble-ai/go-jnode"

func RemoveJNodeElementsIf(n *jnode.Node, f func(*jnode.Node) bool) *jnode.Node {
	// don't allocate a new array if the filter doesn't exclude anything
	var result *jnode.Node
	for i, e := range n.Elements() {
		if f(e) {
			if result == nil {
				result = jnode.NewArrayNode()
				for j := 0; j < i; j++ {
					result.Append(n.Get(j))
				}
			}
		} else if result != nil {
			result.Append(e)
		}
	}
	if result != nil {
		return result
	}
	return n
}
