package util

import "testing"

func TestStringSet(t *testing.T) {
	s := NewStringSet().AddAll("1", "two", "three")
	if s.Add("1") != false {
		t.Error(s)
	}
	if s.Add("one") != true {
		t.Error(s)
	}
	v := s.Values()
	if len(v) != 4 || v[0] != "1" || v[1] != "two" || v[3] != "one" {
		t.Error(v)
	}
	if s.Contains("one") != true || s.Contains("four") == true {
		t.Error(s)
	}
}
