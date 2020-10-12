package scanner

// ResourceConfig describes a resource present in IaC
type ResourceConfig struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Source string      `json:"source"`
	Line   int         `json:"line"`
	Type   string      `json:"type"`
	Config interface{} `json:"config"`
}

// AllResourceConfigs is a list/slice of resource configs present in IaC
type AllResourceConfigs map[string][]ResourceConfig
