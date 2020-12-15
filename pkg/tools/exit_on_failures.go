package tools

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/exit"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/util"
)

var serverityNames = util.NewStringSetWithValues([]string{
	"info", "low", "medium", "high", "critical",
})

func ExitOnFailures(thresholds map[string]string, result *jnode.Node) {
	findings := result.Path("findings")
	parsedThresholds, err := parseFailThresholds(thresholds)
	if err != nil {
		log.Warnf("Ignoring some --fail thresholds: {warning:%s}", err.Error())
	}
	for level, value := range parsedThresholds {
		count := 0
		for _, n := range findings.Elements() {
			if strings.ToLower(n.Path("severity").AsText()) == level {
				if !n.Path("passed").AsBool() {
					count++
				}
			}
		}
		if value > 0 && count > value {
			exit.Code = 2
			exit.Message = fmt.Sprintf("Found {danger:%d failed %s} findings", count, level)
			break
		}
	}
}

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
