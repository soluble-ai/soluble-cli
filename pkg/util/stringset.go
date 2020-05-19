package util

type StringSet struct {
	values []string
	set    map[string]interface{}
}

func NewStringSet() *StringSet {
	return &StringSet{
		set: map[string]interface{}{},
	}
}

func (ss *StringSet) Contains(s string) bool {
	_, ok := ss.set[s]
	return ok
}

// Adds s to the set and returns true if the string wasn't
// already present
func (ss *StringSet) Add(s string) bool {
	_, ok := ss.set[s]
	if !ok {
		ss.set[s] = nil
		ss.values = append(ss.values, s)
	}
	return !ok
}

func (ss *StringSet) AddAll(values ...string) *StringSet {
	for _, s := range values {
		ss.set[s] = nil
	}
	ss.values = append(ss.values, values...)
	return ss
}

func (ss *StringSet) Values() []string {
	return ss.values
}
