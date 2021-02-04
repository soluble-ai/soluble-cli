package assessments

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/go-multierror"
)

func ParseFailThresholds(thresholds []string) (map[string]int, error) {
	var err error
	thresholdsMap := map[string]string{}
	for _, t := range thresholds {
		equals := strings.Index(t, "=")
		var key, val string
		switch {
		case equals == 0:
			err = multierror.Append(err, fmt.Errorf("threshold must be in form severity=count not %s", t))
		case equals > 0:
			val = t[equals+1:]
			key = strings.ToLower(t[0:equals])
		case equals < 0:
			val = "1"
			key = strings.ToLower(t)
		}
		thresholdsMap[key] = val
	}
	last := -1
	result := map[string]int{}
	for _, name := range SeverityNames.Values() {
		if s, ok := thresholdsMap[name]; ok {
			value, convErr := strconv.Atoi(s)
			switch {
			case convErr != nil:
				err = multierror.Append(err, fmt.Errorf("invalid threshold %s for %s", s, name))
			case value == 0:
				err = multierror.Append(err, fmt.Errorf("threshold count for %s must be > 0", name))
			default:
				result[name] = value
				last = value
				continue
			}
		}
		result[name] = last
	}
	for key := range thresholdsMap {
		if !SeverityNames.Contains(key) {
			err = multierror.Append(err, fmt.Errorf("unrecognized level: %s", key))
		}
	}
	return result, err
}
