package util

import (
	"fmt"
	"io/ioutil"
	"testing"
)

type errorCloser struct {
	m      string
	closed bool
}

func (ec *errorCloser) Close() error {
	ec.closed = true
	return fmt.Errorf("%s", ec.m)
}

func TestCloseAll(t *testing.T) {
	ec2 := &errorCloser{"second", false}
	ec3 := &errorCloser{"third", false}
	err := CloseAll(ioutil.NopCloser(nil), ec2, ec3)
	if err == nil {
		t.Fatal(err)
	}
	if err.Error() != "second" {
		t.Error(err)
	}
	if !ec2.closed || !ec3.closed {
		t.Error("not closed")
	}
}

func TestPropagateCloseError(t *testing.T) {
	ec := &errorCloser{"ec", false}
	err := PropagateCloseError(ec, func() error {
		return nil
	})
	if err == nil || err.Error() != "ec" || !ec.closed {
		t.Fatal("bad error", err)
	}
	ec.closed = false
	err = PropagateCloseError(ec, func() error {
		return fmt.Errorf("func")
	})
	if err == nil || err.Error() != "func" || !ec.closed {
		t.Fatal("bad error", err)
	}
}
