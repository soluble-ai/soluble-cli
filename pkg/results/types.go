package results

// Violation Contains data for each violation
type Violation struct {
	RuleName     string      `json:"ruleName" yaml:"rule_name" xml:"rule_name,attr"`
	Description  string      `json:"description" yaml:"description" xml:"description,attr"`
	RuleID       string      `json:"ruleId" yaml:"rule_id" xml:"rule_id,attr"`
	Severity     string      `json:"severity" yaml:"severity" xml:"severity,attr"`
	Category     string      `json:"category" yaml:"category" xml:"category,attr"`
	RuleFile     string      `json:"-" yaml:"-" xml:"-"`
	RuleData     interface{} `json:"-" yaml:"-" xml:"-"`
	ResourceName string      `json:"resourceName" yaml:"resource_name" xml:"resource_name,attr"`
	ResourceType string      `json:"resourceType" yaml:"resource_type" xml:"resource_type,attr"`
	ResourceData interface{} `json:"-" yaml:"-" xml:"-"`
	File         string      `json:"file" yaml:"file" xml:"file,attr"`
	LineNumber   int         `json:"line" yaml:"line" xml:"line,attr"`
}

// ViolationStats Contains stats related to the violation data
type ViolationStats struct {
	LowCount    int `json:"low" yaml:"low" xml:"low,attr"`
	MediumCount int `json:"medium" yaml:"medium" xml:"medium,attr"`
	HighCount   int `json:"high" yaml:"high" xml:"high,attr"`
	TotalCount  int `json:"total" yaml:"total" xml:"total,attr"`
}

// ViolationStore Storage area for violation data
type ViolationStore struct {
	Violations []*Violation   `json:"violations" yaml:"violations" xml:"violations>violation"`
	Count      ViolationStats `json:"count" yaml:"count" xml:"count"`
}
