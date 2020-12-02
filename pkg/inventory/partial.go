package inventory

import (
	"strings"

	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
)

func PartialDecodeJSON(buf []byte) map[string]string {
	r := map[string]string{}
	// gjson doesn't require valid JSON
	for k, v := range gjson.ParseBytes(buf).Map() {
		if v.Type == gjson.String {
			r[k] = v.Str
		}
	}
	return r
}

func PartialDecodeYAML(buf []byte) map[string]string {
	var m map[string]interface{}
	// truncated yaml is still mostly yaml, but we'll just ignore errors
	_ = yaml.Unmarshal(buf, &m)
	r := map[string]string{}
	for k, v := range m {
		if s, ok := v.(string); ok {
			r[k] = s
		}
	}
	return r
}

func PartialDecode(name string, buf []byte) map[string]string {
	switch {
	default:
		return PartialDecodeJSON(buf)
	case strings.HasSuffix(name, ".yaml"):
		return PartialDecodeYAML(buf)
	}
}
