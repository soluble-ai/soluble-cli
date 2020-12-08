package iacscan

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/exit"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/model"
)

func ExitOnFailures(command model.Command, result *jnode.Node) (*jnode.Node, error) {
	findings := result.Path("findings")
	thresholds, _ := command.GetCobraCommand().Flags().GetStringToString("fail")
	for level, s := range thresholds {
		value, err := strconv.Atoi(s)
		if err != nil {
			log.Warnf("Ignoring failure count {warning:%s} for level {warning:%s}", s, level)
			continue
		}
		count := 0
		for _, n := range findings.Elements() {
			if strings.ToLower(n.Path("severity").AsText()) == level {
				if n.Path("passed").AsText() != "true" {
					count++
				}
			}
		}
		if count > value {
			exit.Code = 2
			exit.Message = fmt.Sprintf("Found {danger:%d failed %s} findings", count, level)
			break
		}
	}
	return result, nil
}
