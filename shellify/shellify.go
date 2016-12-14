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
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/zombiezen/mcm/catalog"
	"github.com/zombiezen/mcm/shellify/shlib"
	"github.com/zombiezen/mcm/third_party/golang/capnproto"
)

func init() {
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [CATALOG]\n", os.Args[0])
	// TODO(someday): flag.PrintDefaults()
}

func main() {
	flag.Parse()

	c, err := readCatalogArg()
	if err != nil {
		fmt.Fprintln(os.Stderr, "mcm-shellify: read catalog:", err)
		os.Exit(1)
	}
	if err = shlib.WriteScript(os.Stdout, c); err != nil {
		fmt.Fprintln(os.Stderr, "mcm-shellify:", err)
		os.Exit(1)
	}
}

func readCatalogArg() (catalog.Catalog, error) {
	switch flag.NArg() {
	case 0:
		return readCatalog(os.Stdin)
	case 1:
		// TODO(someday): read segments lazily
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			return catalog.Catalog{}, err
		}
		defer f.Close()
		return readCatalog(f)
	default:
		usage()
		os.Exit(2)
		panic("unreachable")
	}
}

func readCatalog(r io.Reader) (catalog.Catalog, error) {
	msg, err := capnp.NewDecoder(r).Decode()
	if err != nil {
		return catalog.Catalog{}, err
	}
	return catalog.ReadRootCatalog(msg)
}
