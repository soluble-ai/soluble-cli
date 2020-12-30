package tools

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/go-multierror"
	"github.com/soluble-ai/soluble-cli/pkg/util"
)

var serverityNames = util.NewStringSetWithValues([]string{
	"info", "low", "medium", "high", "critical",
})

func parseFailThresholds(thresholds map[string]string) (map[string]int, error) {
	last := -1
	result := map[string]int{}
	var err error
	for _, name := range serverityNames.Values() {
		if s, ok := thresholds[name]; ok {
			value, convErr := strconv.Atoi(s)
			if convErr != nil {
				err = multierror.Append(err, fmt.Errorf("invalid threshold %s for %s", s, name))
			} else {
				result[name] = value
				last = value
				continue
			}
		}
		result[name] = last
	}
	for key := range thresholds {
		if !serverityNames.Contains(key) {
			err = multierror.Append(err, fmt.Errorf("unrecognized level: %s", key))
		}
	}
	return result, err
}
