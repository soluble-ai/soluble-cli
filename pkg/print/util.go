package print

import (
	"encoding/json"

	"github.com/soluble-ai/go-jnode"
)

// ToResult converts a value to a jnode.Node
func ToResult(value interface{}) (*jnode.Node, error) {
	dat, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return jnode.FromJSON(dat)
}
