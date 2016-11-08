package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/zombiezen/mcm/catalog"
	"github.com/zombiezen/mcm/internal/depgraph"
	"github.com/zombiezen/mcm/third_party/golang/capnproto"
)

func main() {
	c, err := readCatalog(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "mcm-exec: read catalog:", err)
		os.Exit(1)
	}
	res, _ := c.Resources()
	g, err := depgraph.New(res)
	if err != nil {
		fmt.Fprintln(os.Stderr, "mcm-exec:", err)
		os.Exit(1)
	}
	if err = applyCatalog(g); err != nil {
		fmt.Fprintln(os.Stderr, "mcm-exec:", err)
		os.Exit(1)
	}
}

func applyCatalog(g *depgraph.Graph) error {
	ok := true
	for !g.Done() {
		ready := g.Ready()
		if len(ready) == 0 {
			return errors.New("graph not done, but has nothing to do")
		}
		curr := ready[0]
		if err := applyResource(g.Resource(curr)); err == nil {
			g.Mark(curr)
		} else {
			// TODO(soon): log skipped resources
			g.MarkFailure(curr)
			fmt.Fprintf(os.Stderr, "mcm-exec: applying resource %d:", curr, err)
			ok = false
		}
	}
	if !ok {
		return errors.New("not all resources applied cleanly")
	}
	return nil
}

func applyResource(r catalog.Resource) error {
	f, _ := r.File()
	path, err := f.Path()
	if err != nil {
		return fmt.Errorf("reading file path: %v", err)
	} else if path == "" {
		return errors.New("file path is empty")
	}
	fmt.Printf("considering file %s\n", path)
	switch f.Which() {
	case catalog.File_Which_plain:
		if f.Plain().HasContent() {
			content, _ := f.Plain().Content()
			// TODO(soon): respect file mode
			// TODO(soon): collect errors instead of stopping
			fmt.Printf("writing file %s\n", path)
			if err := ioutil.WriteFile(path, content, 0666); err != nil {
				return err
			}
		}
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
