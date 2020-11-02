package cloudformationguard

import (
	"bufio"
	"regexp"
	"strings"
)

type failure struct {
	Resource       string `json:"resource"`
	Attribute      string `json:"attribute"`
	AttributeValue string `json:"attribute_value"`
	Message        string `json:"message"`
}

var (
	failedPattern = regexp.MustCompile(`\[(\w+)] failed because \[([^]]+)] is \[([^]]+)] and (.*)`)
)

func parseFailures(output string) []failure {
	result := []failure{}
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		m := failedPattern.FindStringSubmatch(line)
		if m != nil {
			result = append(result, failure{
				Resource:       m[1],
				Attribute:      m[2],
				AttributeValue: m[3],
				Message:        m[4],
			})
		}
	}
	return result
}
