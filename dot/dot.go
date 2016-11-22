package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/zombiezen/mcm/catalog"
	"github.com/zombiezen/mcm/third_party/golang/capnproto"
)

func main() {
	flag.Parse()

	var cat catalog.Catalog
	switch flag.NArg() {
	case 0:
		var err error
		cat, err = readCatalog(os.Stdin)
		if err != nil {
			die(err)
		}
	case 1:
		// TODO(someday): read segments lazily
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			die(err)
		}
		cat, err = readCatalog(f)
		if err != nil {
			die(err)
		}
		if err = f.Close(); err != nil {
			die(err)
		}
	default:
		flag.Usage()
		os.Exit(2)
	}

	fmt.Println("digraph catalog {")
	resources, _ := cat.Resources()
	for i := 0; i < resources.Len(); i++ {
		r := resources.At(i)
		id := r.ID()
		if c, _ := r.Comment(); c != "" {
			fmt.Printf("  %d [label=%q];\n", id, c)
		}
		deps, _ := r.Dependencies()
		for j := 0; j < deps.Len(); j++ {
			fmt.Printf("  %d -> %d;\n", id, deps.At(j))
		}
		fmt.Println()
	}
	fmt.Println("}")
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "mcm-dot:", err)
	os.Exit(1)
}

func readCatalog(r io.Reader) (catalog.Catalog, error) {
	msg, err := capnp.NewDecoder(r).Decode()
	if err != nil {
		return catalog.Catalog{}, fmt.Errorf("read catalog: %v", err)
	}
	c, err := catalog.ReadRootCatalog(msg)
	if err != nil {
		return catalog.Catalog{}, fmt.Errorf("read catalog: %v", err)
	}
	return c, nil
}
