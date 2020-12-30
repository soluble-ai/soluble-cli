package tools

import "strings"

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

func (a *Assessment) HasFailures(thresholds map[string]int) (bool, int, string) {
	counts := map[string]int{}
	for _, f := range a.Findings {
		counts[strings.ToLower(f.Severity)] += 1
	}
	for _, level := range serverityNames.Values() {
		value := thresholds[level]
		count := counts[level]
		if value > 0 && count >= value {
			return true, count, level
		}
	}
	return false, 0, ""
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
