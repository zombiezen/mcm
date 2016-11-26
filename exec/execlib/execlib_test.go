package execlib

import (
	"bytes"
	"context"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/zombiezen/mcm/catalog"
	"github.com/zombiezen/mcm/internal/system/fakesystem"
	"github.com/zombiezen/mcm/third_party/golang/capnproto"
	"github.com/zombiezen/mcm/third_party/golang/capnproto/pogs"
)

func TestEmpty(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cat, err := new(catalogStruct).toCapnp()
	if err != nil {
		t.Fatal("new(catalogStruct).toCapnp():", err)
	}
	app := &Applier{
		System: new(fakesystem.System),
		Log:    testLogger{t: t},
	}
	err = app.Apply(ctx, cat)
	if err != nil {
		t.Error("Apply:", err)
	}
}

func TestFile(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fpath := filepath.Join(fakesystem.Root, "foo")
	cat, err := (&catalogStruct{
		Resources: []resource{
			{
				ID:      42,
				Comment: "file",
				Which:   catalog.Resource_Which_file,
				File:    newPlainFile(fpath, []byte("Hello")),
			},
		},
	}).toCapnp()
	if err != nil {
		t.Fatal("catalogStruct.toCapnp():", err)
	}
	sys := new(fakesystem.System)
	app := &Applier{
		System: sys,
		Log:    testLogger{t: t},
	}
	err = app.Apply(ctx, cat)
	if err != nil {
		t.Error("Apply:", err)
	}

	f, err := sys.OpenFile(ctx, fpath)
	if err != nil {
		t.Fatalf("sys.OpenFile(ctx, %q): %v", fpath, err)
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		t.Errorf("ioutil.ReadAll(sys.OpenFile(ctx, %q)): %v", fpath, err)
	}
	if !bytes.Equal(data, []byte("Hello")) {
		t.Errorf("ioutil.ReadAll(sys.OpenFile(ctx, %q)) = %q; want \"Hello\"", fpath, data)
	}
}

type catalogStruct struct {
	Resources []resource
}

func (c *catalogStruct) toCapnp() (catalog.Catalog, error) {
	_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return catalog.Catalog{}, err
	}
	root, err := catalog.NewRootCatalog(seg)
	if err != nil {
		return catalog.Catalog{}, err
	}
	err = pogs.Insert(catalog.Catalog_TypeID, root.Struct, c)
	return root, err
}

type resource struct {
	ID      uint64 `capnp:"id"`
	Comment string
	Deps    []uint64 `capnp:"dependencies"`

	Which catalog.Resource_Which
	File  *file
}

type file struct {
	Path string

	Which catalog.File_Which
	Plain struct {
		Content []byte
	}
	Directory struct{}
	Symlink   struct {
		Target string
	}
}

func newPlainFile(path string, content []byte) *file {
	f := &file{Path: path, Which: catalog.File_Which_plain}
	f.Plain.Content = content
	return f
}

type testLogger struct {
	t interface {
		Logf(string, ...interface{})
	}
}

func (tl testLogger) Infof(ctx context.Context, format string, args ...interface{}) {
	tl.t.Logf("applier info: "+format, args...)
}

func (tl testLogger) Error(ctx context.Context, err error) {
	tl.t.Logf("applier error: %v", err)
}
