package assessments

import (
	"encoding/json"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/soluble-ai/soluble-cli/pkg/api"
	"github.com/soluble-ai/soluble-cli/pkg/util"
	"github.com/soluble-ai/soluble-cli/pkg/xcp"
)

type Assessment struct {
	ID       string `json:"assessmentId"`
	URL      string `json:"appUrl"`
	Title    string
	Module   string
	Category string
	Markdown string
	Findings []Finding
	Failed   bool `json:"-"`
}

type Assessments []Assessment

type Finding struct {
	SID         string
	Severity    string
	Title       string
	Description string
	Markdown    string
	FilePath    string `json:"filePath"`
	Line        int
	Pass        bool
}

var SeverityNames = util.NewStringSetWithValues([]string{
	"info", "low", "medium", "high", "critical",
})

func (a *Assessment) HasFailures(thresholds map[string]int) (bool, string, int) {
	counts := map[string]int{}
	for _, f := range a.Findings {
		counts[strings.ToLower(f.Severity)] += 1
	}
	for _, level := range SeverityNames.Values() {
		value := thresholds[level]
		count := counts[level]
		if value > 0 && count >= value {
			return true, level, count
		}
	}
	return false, "", 0
}

func (f *Finding) GetTitle() string {
	if f.Title != "" {
		return f.Title
	}
	t := f.Description
	const lim = 57
	if len(t) > lim+3 {
		t = t[0:lim] + "..."
	}
	return t
}

func FindCIEnvAssessments(client *api.Client) (Assessments, error) {
	n, err := client.Get("/api/v1/org/{org}/assessments",
		func(r *resty.Request) {
			r.SetQueryParam("detail", "true")
			r.SetQueryParam("searchType", "ci")
		},
		xcp.WithCIEnv)
	if err != nil {
		return nil, err
	}
	assessments := []Assessment{}
	for _, n := range n.Path("data").Elements() {
		assessment := Assessment{}
		if err := json.Unmarshal([]byte(n.String()), &assessment); err != nil {
			return nil, err
		}
		assessments = append(assessments, assessment)
	}
	return assessments, nil
}
