package login

import "testing"

func TestMakeState(t *testing.T) {
	s1 := MakeState()
	if len(s1) != 64 {
		t.Error(s1)
	}
	if s2 := MakeState(); len(s2) != 64 || s1 == s2 {
		t.Error(s1, s2)
	}
}
