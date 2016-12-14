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
	"errors"
	"fmt"
	"io"
	"strconv"
)

type gen struct {
	ew     errWriter
	indent int
}

func newGen(w io.Writer) *gen {
	return &gen{ew: errWriter{w: w}}
}

var newline = []byte{'\n'}

func (g *gen) p(args ...interface{}) {
	if len(args) == 0 {
		g.ew.Write(newline)
		return
	}
	var buf []byte
	for i := 0; i < g.indent; i++ {
		buf = append(buf, '\t')
	}
	for i, a := range args {
		if i > 0 {
			buf = append(buf, ' ')
		}
		if a, ok := a.(heredoc); ok {
			if i != len(args)-1 {
				panic(errors.New("heredoc placed in non-final argument"))
			}
			g.ew.Write(buf)
			a.encode(&g.ew)
			g.ew.Write(newline)
			return
		}
		buf = appendPArg(buf, a)
	}
	buf = append(buf, '\n')
	g.ew.Write(buf)
}

func appendPArg(buf []byte, a interface{}) []byte {
	switch a := a.(type) {
	case string:
		return appendShellQuote(buf, a)
	case script:
		return append(buf, a...)
	case int:
		return strconv.AppendInt(buf, int64(a), 10)
	case assignment:
		// TODO(maybe): check name for shell safety?
		buf = append(buf, a.name...)
		buf = append(buf, '=')
		return appendPArg(buf, a.value)
	default:
		panic(fmt.Errorf("unknown type: %T", a))
	}
}

func (g *gen) in()  { g.indent++ }
func (g *gen) out() { g.indent-- }

type errWriter struct {
	w   io.Writer
	err error
}

func newErrWriter(w io.Writer) *errWriter {
	if ew, ok := w.(*errWriter); ok {
		return ew
	}
	return &errWriter{w: w}
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

type heredoc struct {
	marker string
	// TODO(someday): use io.WriterTo to avoid copying in base64 case
	data []byte
}

func (hd heredoc) encode(w io.Writer) error {
	ew := newErrWriter(w)
	ew.WriteString("<<'")
	ew.WriteString(hd.marker)
	ew.WriteString("'\n")
	ew.Write(hd.data)
	ew.Write(newline)
	ew.WriteString(hd.marker)
	return ew.err
}

func (hd heredoc) String() string {
	var buf bytes.Buffer
	hd.encode(&buf)
	return buf.String()
}

type assignment struct {
	name  string
	value interface{}
}

func (a assignment) String() string {
	return string(appendPArg(nil, a))
}

func appendShellQuote(buf []byte, s string) []byte {
	if s == "" {
		return append(buf, '\'', '\'')
	}
	safe := true
	for i := 0; i < len(s); i++ {
		if !isShellSafe(s[i]) {
			safe = false
			break
		}
	}
	if safe {
		return append(buf, s...)
	}
	buf = append(buf, '\'')
	for i := 0; i < len(s); i++ {
		if s[i] == '\'' {
			buf = append(buf, '\'', '\\', '\'', '\'')
		} else {
			buf = append(buf, s[i])
		}
	}
	buf = append(buf, '\'')
	return buf
}

func isShellSafe(b byte) bool {
	return b >= 'A' && b <= 'Z' || b >= 'a' && b <= 'z' || b >= '0' && b <= '9' || b == '-' || b == '_' || b == '/' || b == '.'
}
