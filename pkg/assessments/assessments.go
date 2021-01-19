package assessments

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/soluble-ai/soluble-cli/pkg/api"
	"github.com/soluble-ai/soluble-cli/pkg/assessments/fingerprint"
	"github.com/soluble-ai/soluble-cli/pkg/log"
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
	Findings []*Finding

	Failed         bool
	FailedCount    int
	FailedSeverity string
}

type Assessments []*Assessment

type Finding struct {
	SID                string            `json:"sid,omitempty"`
	Severity           string            `json:"severity,omitempty"`
	Title              string            `json:"title,omitempty"`
	Description        string            `json:"description,omitempty"`
	Markdown           string            `json:"markdown,omitempty"`
	FilePath           string            `json:"filePath,omitempty"`
	Line               int               `json:"line,omitempty"`
	Pass               bool              `json:"pass,omitempty"`
	PartialFingerprint string            `json:"partialFingerprint,omitempty"`
	Tool               map[string]string `json:"tool,omitempty"`
}

var SeverityNames = util.NewStringSetWithValues([]string{
	"info", "low", "medium", "high", "critical",
})

func (a *Assessment) EvaluateFailures(thresholds map[string]int) {
	counts := map[string]int{}
	for _, f := range a.Findings {
		counts[strings.ToLower(f.Severity)] += 1
	}
	for _, level := range SeverityNames.Values() {
		value := thresholds[level]
		count := counts[level]
		if value > 0 && count >= value {
			a.Failed = true
			a.FailedSeverity = level
			a.FailedCount = count
			return
		}
	}
}

func ComputePartialFingerprints(dir string, findings []*Finding) {
	findingsForFiles := map[string][]*Finding{}
	for _, f := range findings {
		if f.FilePath != "" && f.Line > 0 {
			findingsForFiles[f.FilePath] = append(findingsForFiles[f.FilePath], f)
		}
	}
	for filePath, fs := range findingsForFiles {
		file, err := os.Open(filepath.Join(dir, filePath))
		if err != nil {
			log.Warnf("Could not read {warning:%s} - %s", filePath, err.Error())
			continue
		}
		defer file.Close()
		findingsForLine := map[int][]*Finding{}
		for _, f := range fs {
			findingsForLine[f.Line] = append(findingsForLine[f.Line], f)
		}
		err = fingerprint.Partial(bufio.NewReader(file), func(lineNumber int, fingerprint string) {
			for _, f := range findingsForLine[lineNumber] {
				f.PartialFingerprint = fingerprint
			}
		})
		if err != nil {
			log.Warnf("Could not compute partial fingerprint for %s - %s", filePath, err.Error())
		}
	}
}

func (f *Finding) SetAttribute(name, value string) *Finding {
	if f.Tool == nil {
		f.Tool = map[string]string{}
	}
	f.Tool[name] = value
	return f
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
		xcp.WithCIEnv(""))
	if err != nil {
		return nil, err
	}
	assessments := []*Assessment{}
	for _, n := range n.Path("data").Elements() {
		assessment := &Assessment{}
		if err := json.Unmarshal([]byte(n.String()), assessment); err != nil {
			return nil, err
		}
		assessments = append(assessments, assessment)
	}
	return assessments, nil
}
