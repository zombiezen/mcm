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
	"bytes"
	"fmt"
	"io"
	"strconv"
)

type gen struct {
	ew     errWriter
	indent int
}

func (g *gen) p(args ...interface{}) {
	if len(args) == 0 {
		g.ew.Write([]byte{'\n'})
		return
	}
	var buf bytes.Buffer
	for i := 0; i < g.indent; i++ {
		buf.WriteString("  ")
	}
	for _, a := range args {
		switch a := a.(type) {
		case string:
			buf.WriteString(shellQuote(a))
		case script:
			buf.WriteString(string(a))
		case uint64:
			buf.WriteString(strconv.FormatUint(a, 10))
		default:
			panic(fmt.Errorf("unknown type: %T", a))
		}
	}
	buf.WriteByte('\n')
	buf.WriteTo(&g.ew)
}

func (g *gen) in()  { g.indent++ }
func (g *gen) out() { g.indent-- }

type errWriter struct {
	w   io.Writer
	err error
}

func (ew *errWriter) Write(p []byte) (n int, err error) {
	if ew.err != nil {
		return 0, ew.err
	}
	n, ew.err = ew.w.Write(p)
	return n, ew.err
}

func (ew *errWriter) WriteString(s string) (n int, err error) {
	if ew.err != nil {
		return 0, ew.err
	}
	n, ew.err = io.WriteString(ew.w, s)
	return n, ew.err
}

// script is properly escaped bash.
type script string

func (s script) String() string {
	return string(s)
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	safe := true
	for i := 0; i < len(s); i++ {
		if !isShellSafe(s[i]) {
			safe = false
			break
		}
	}
	if safe {
		return s
	}
	buf := make([]byte, 0, len(s)+2)
	buf = append(buf, '\'')
	for i := 0; i < len(s); i++ {
		if s[i] == '\'' {
			buf = append(buf, '\'', '\\', '\'', '\'')
		} else {
			buf = append(buf, s[i])
		}
	}
	buf = append(buf, '\'')
	return string(buf)
}

func isShellSafe(b byte) bool {
	return b >= 'A' && b <= 'Z' || b >= 'a' && b <= 'z' || b >= '0' && b <= '9' || b == '-' || b == '_' || b == '/'
}
