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

package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/zombiezen/mcm/catalog"
	"github.com/zombiezen/mcm/internal/depgraph"
	"github.com/zombiezen/mcm/third_party/golang/capnproto"
)

func main() {
	c, err := readCatalog(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "mcm-shellify: read catalog:", err)
		os.Exit(1)
	}
	res, _ := c.Resources()
	g, err := depgraph.New(res)
	if err != nil {
		fmt.Fprintln(os.Stderr, "mcm-shellify:", err)
		os.Exit(1)
	}
	if err = writeScript(os.Stdout, g); err != nil {
		fmt.Fprintln(os.Stderr, "mcm-shellify:", err)
		os.Exit(1)
	}
}

func writeScript(w io.Writer, g *depgraph.Graph) error {
	ew := &errWriter{w: w}
	io.WriteString(ew, "#!/bin/bash\n_() {\nset -e\n")
	for ew.err == nil && !g.Done() {
		ready := append([]uint64(nil), g.Ready()...)
		if len(ready) == 0 {
			return errors.New("graph not done, but has nothing to do")
		}
		for _, id := range ready {
			if err := scriptResource(ew, g.Resource(id)); err != nil {
				return fmt.Errorf("resource ID=%d: %v", id, err)
			}
			g.Mark(id)
		}
	}
	io.WriteString(ew, "}\n_ \"$0\" \"$@\"\n")
	return ew.err
}

func scriptResource(ew *errWriter, r catalog.Resource) error {
	if c, _ := r.Comment(); c != "" {
		fmt.Fprintf(ew, "\n# %s\n", c)
	} else {
		fmt.Fprintf(ew, "\n# Resource ID=%d\n", r.ID())
	}
	f, _ := r.File()
	path, err := f.Path()
	if err != nil {
		return fmt.Errorf("reading file path: %v", err)
	} else if path == "" {
		return errors.New("file path is empty")
	}
	switch f.Which() {
	case catalog.File_Which_plain:
		// TODO(soon): touch, even if no content
		// TODO(soon): respect file mode
		if f.Plain().HasContent() {
			fmt.Fprintf(ew, "base64 -d > %s <<!EOF!\n", shellQuote(path))
			content, _ := f.Plain().Content()
			enc := base64.NewEncoder(base64.StdEncoding, ew)
			enc.Write(content)
			enc.Close()
			io.WriteString(ew, "\n!EOF!\n")
		}
	case catalog.File_Which_directory:
		// TODO(soon): respect file mode
		fmt.Fprintf(ew, "if [[ ! -d %s ]]; then\n\tmkdir %[1]s\nfi\n", shellQuote(path))
	default:
		return fmt.Errorf("unsupported file directive %v", f.Which())
	}
	return nil
}

func readCatalog(r io.Reader) (catalog.Catalog, error) {
	msg, err := capnp.NewDecoder(r).Decode()
	if err != nil {
		return catalog.Catalog{}, err
	}
	return catalog.ReadRootCatalog(msg)
}

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

func shellQuote(s string) string {
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
