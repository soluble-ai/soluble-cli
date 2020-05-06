package model

import (
	"fmt"
)

// What to do with a parameter.  By default the parameter is put
// in the query string (for GET), or in the json body (everything else.)
type ParameterDisposition string

const (
	// Put the parameter value in the context
	ContextDisposition = ParameterDisposition("context")
)

func (d ParameterDisposition) validate() error {
	if d == "" || d == ContextDisposition {
		return nil
	}
	return fmt.Errorf("invalid parameter disposition '%s' must be one of %s",
		d, ContextDisposition)
}
