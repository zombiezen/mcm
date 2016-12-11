// Copyright 2016 The Minimal Configuration Manager Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package shlib

import (
	"testing"
)

func TestShellQuote(t *testing.T) {
	tests := []struct {
		in, out string
	}{
		{``, `''`},
		{`abc`, `abc`},
		{`abc def`, `'abc def'`},
		{`abc/def`, `abc/def`},
		{`"abc"`, `'"abc"'`},
		{`'abc'`, `''\''abc'\'''`},
		{`abc\`, `'abc\'`},
	}
	t.Run("NilAppend", func(t *testing.T) {
		for _, test := range tests {
			out := string(appendShellQuote(nil, test.in))
			if out != test.out {
				t.Errorf("shellQuote(%q) = %s; want %s", test.in, out, test.out)
			}
		}
	})
	t.Run("Prefix", func(t *testing.T) {
		for _, test := range tests {
			out := string(appendShellQuote([]byte("AAA"), test.in))
			if want := "AAA" + test.out; out != want {
				t.Errorf("shellQuote(%q) = %s; want %s", test.in, out, want)
			}
		}
	})
}
