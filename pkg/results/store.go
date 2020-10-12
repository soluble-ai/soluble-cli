package results

// NewViolationStore returns a new violation store
func NewViolationStore() *ViolationStore {
	return &ViolationStore{
		Violations: []*Violation{},
	}
}

// AddResult Adds individual violations into the violation store
func (s *ViolationStore) AddResult(violation *Violation) {
	s.Violations = append(s.Violations, violation)
}

// GetResults Retrieves all violations from the violation store
func (s *ViolationStore) GetResults() []*Violation {
	return s.Violations
}
