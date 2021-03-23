// Copyright 2021 Soluble Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
