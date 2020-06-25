// Copyright 2020 Soluble Inc
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

package print

import (
	"testing"
	"time"

	"github.com/soluble-ai/go-jnode"
)

func TestTimestampFormatters(t *testing.T) {
	now := time.Date(2020, 5, 11, 10, 5, 45, 0, time.UTC)
	formatterNow = &now
	formatterLocation = time.FixedZone("test", -8*60*60)
	n := jnode.NewObjectNode().Put("ts", "2020-05-08T11:18:33Z")
	if s := TimestampFormatter(n, "ts"); s != "2020-05-08T03:18:33-08:00" {
		t.Error("timestamp wrong", n, s)
	}
	if s := RelativeTimestampFormatter(n, "ts"); s != "2d22h47m12s" {
		t.Error("relative ts wrong", n, s)
	}
}

var bytesTestCases = []struct {
	value    int
	expected string
}{
	{10, "10"},
	{1536, "1.5K"},
	{1258291, "1.2M"},
	{2684354560, "2.5G"},
}

func TestBytesFormatter(t *testing.T) {
	for _, c := range bytesTestCases {
		n := jnode.NewObjectNode().Put("n", c.value)
		s := BytesFormatter(n, "n")
		if s != c.expected {
			t.Error(c.value, c.expected, s)
		}
	}
}
