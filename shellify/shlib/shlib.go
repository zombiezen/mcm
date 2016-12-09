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

// Package shlib provides the functionality of the mcm-shellify tool.
package shlib

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"github.com/zombiezen/mcm/catalog"
	"github.com/zombiezen/mcm/internal/depgraph"
)

// WriteScript converts a catalog into a bash script and writes it to w.
func WriteScript(w io.Writer, c catalog.Catalog) error {
	res, _ := c.Resources()
	graph, err := depgraph.New(res)
	if err != nil {
		return err
	}
	g := newGen(w)
	g.p(script("#!/bin/bash"))
	g.p(script("_() {"))
	g.in()
	g.p(script("set -e"))
	for g.ew.err == nil && !graph.Done() {
		ready := append([]uint64(nil), graph.Ready()...)
		if len(ready) == 0 {
			return errors.New("graph not done, but has nothing to do")
		}
		for _, id := range ready {
			if err := g.resource(graph.Resource(id)); err != nil {
				return fmt.Errorf("resource ID=%d: %v", id, err)
			}
			graph.Mark(id)
		}
	}
	g.out()
	g.p(script("}"))
	g.p(script(`_ "$0" "$@"`))
	return g.ew.err
}

func (g *gen) resource(r catalog.Resource) error {
	if r.Which() == catalog.Resource_Which_noop {
		return nil
	}
	g.p()
	if c, _ := r.Comment(); c != "" {
		g.p(script("#"), script(c))
	} else {
		g.p(script("# Resource ID ="), r.ID())
	}
	switch r.Which() {
	case catalog.Resource_Which_file:
		f, err := r.File()
		if err != nil {
			return fmt.Errorf("read from catalog: %v", err)
		}
		return g.file(f)
	default:
		return fmt.Errorf("unsupported resource %v", r.Which())
	}
}

func (g *gen) file(f catalog.File) error {
	path, err := f.Path()
	if err != nil {
		return fmt.Errorf("reading file path: %v", err)
	} else if path == "" {
		return errors.New("file path is empty")
	}
	switch f.Which() {
	case catalog.File_Which_plain:
		// TODO(soon): handle no content case
		// TODO(soon): respect file mode
		if f.Plain().HasContent() {
			content, err := f.Plain().Content()
			if err != nil {
				return fmt.Errorf("read content from catalog: %v", err)
			}
			enc := make([]byte, base64.StdEncoding.EncodedLen(len(content)))
			base64.StdEncoding.Encode(enc, content)
			g.p(script("base64 -d >"), path, heredoc{marker: "!EOF!", data: enc})
		}
	case catalog.File_Which_directory:
		// TODO(soon): respect file mode
		g.p(script("if [[ ! -d"), path, script("]]; then"))
		g.in()
		g.p(script("mkdir"), path)
		g.out()
		g.p(script("fi"))
	case catalog.File_Which_symlink:
		target, _ := f.Symlink().Target()
		if target == "" {
			return errors.New("symlink target is empty")
		}
		g.p(script("if [[ ! -e"), path, script("]]; then"))
		g.in()
		g.p(script("ln -s"), target, path)
		g.out()
		g.p(script("elif [[ -L"), path, script("]]; then"))
		g.in()
		g.p(script("ln -f -s"), target, path)
		g.out()
		g.p(script("else"))
		g.in()
		// TODO(soon): skip dependent tasks on failure
		g.p(script("echo"), path, script("'is not a symlink' 1>&2"))
		g.p(script("return 1"))
		g.out()
		g.p(script("fi"))
	default:
		return fmt.Errorf("unsupported file directive %v", f.Which())
	}
	return nil
}
