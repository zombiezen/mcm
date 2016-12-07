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

package execlib

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/zombiezen/mcm/catalog"
	. "github.com/zombiezen/mcm/exec/execlib"
	"github.com/zombiezen/mcm/internal/system"
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

func TestLink(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fpath := filepath.Join(fakesystem.Root, "foo")
	lpath := filepath.Join(fakesystem.Root, "link")
	cat, err := (&catalogStruct{
		Resources: []resource{
			{
				ID:      42,
				Comment: "file",
				Which:   catalog.Resource_Which_file,
				File:    newPlainFile(fpath, []byte("Hello")),
			},
			{
				ID:      100,
				Deps:    []uint64{42},
				Comment: "link",
				Which:   catalog.Resource_Which_file,
				File:    newSymlinkFile(fpath, lpath),
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

	if info, err := sys.Lstat(ctx, lpath); err == nil {
		if info.Mode()&os.ModeType != os.ModeSymlink {
			t.Errorf("sys.Lstat(ctx, %q).Mode() = %v; want symlink", lpath, info.Mode())
		}
	} else {
		t.Errorf("sys.Lstat(ctx, %q): %v", lpath, err)
	}
	if target, err := sys.Readlink(ctx, lpath); err == nil {
		if target != fpath {
			t.Errorf("sys.Readlink(ctx, %q) = %q; want %q", lpath, target, fpath)
		}
	} else {
		t.Errorf("sys.Readlink(ctx, %q): %v", lpath, err)
	}
	if f, err := sys.OpenFile(ctx, lpath); err == nil {
		defer f.Close()
		data, err := ioutil.ReadAll(f)
		if err != nil {
			t.Errorf("ioutil.ReadAll(sys.OpenFile(ctx, %q)): %v", lpath, err)
		}
		if !bytes.Equal(data, []byte("Hello")) {
			t.Errorf("ioutil.ReadAll(sys.OpenFile(ctx, %q)) = %q; want \"Hello\"", lpath, data)
		}
	} else {
		t.Errorf("sys.OpenFile(ctx, %q): %v", lpath, err)
	}
}

func TestRelink(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	f1path := filepath.Join(fakesystem.Root, "foo")
	f2path := filepath.Join(fakesystem.Root, "bar")
	lpath := filepath.Join(fakesystem.Root, "link")
	cat, err := (&catalogStruct{
		Resources: []resource{
			{
				ID:      42,
				Comment: "link",
				Which:   catalog.Resource_Which_file,
				File:    newSymlinkFile(f2path, lpath),
			},
		},
	}).toCapnp()
	if err != nil {
		t.Fatal("catalogStruct.toCapnp():", err)
	}
	sys := new(fakesystem.System)
	if err := system.WriteFile(ctx, sys, f1path, []byte("File 1"), 0666); err != nil {
		t.Fatal("WriteFile 1:", err)
	}
	if err := system.WriteFile(ctx, sys, f2path, []byte("File 2"), 0666); err != nil {
		t.Fatal("WriteFile 2:", err)
	}
	if err := sys.Symlink(ctx, f1path, lpath); err != nil {
		t.Fatalf("Symlink %s -> %s: %v", lpath, f1path, err)
	}

	app := &Applier{
		System: sys,
		Log:    testLogger{t: t},
	}
	err = app.Apply(ctx, cat)
	if err != nil {
		t.Error("Apply:", err)
	}

	if info, err := sys.Lstat(ctx, lpath); err == nil {
		if info.Mode()&os.ModeType != os.ModeSymlink {
			t.Errorf("sys.Lstat(ctx, %q).Mode() = %v; want symlink", lpath, info.Mode())
		}
	} else {
		t.Errorf("sys.Lstat(ctx, %q): %v", lpath, err)
	}
	if target, err := sys.Readlink(ctx, lpath); err == nil {
		if target != f2path {
			t.Errorf("sys.Readlink(ctx, %q) = %q; want %q", lpath, target, f2path)
		}
	} else {
		t.Errorf("sys.Readlink(ctx, %q): %v", lpath, err)
	}
}

func TestExec(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	binpath := filepath.Join(fakesystem.Root, "bin")
	aptpath := filepath.Join(binpath, "apt-get")
	cat, err := (&catalogStruct{
		Resources: []resource{
			{
				ID:      42,
				Comment: "apt-get update",
				Which:   catalog.Resource_Which_exec,
				Exec: &exec{
					Command: &command{
						Which: catalog.Exec_Command_Which_argv,
						Argv:  []string{aptpath, "update"},
					},
					Condition: execCondition{Which: catalog.Exec_condition_Which_always},
				},
			},
		},
	}).toCapnp()
	if err != nil {
		t.Fatal("catalogStruct.toCapnp():", err)
	}
	sys := new(fakesystem.System)
	if err := sys.Mkdir(ctx, binpath, 0777); err != nil {
		t.Fatalf("mkdir %s: %v", binpath, err)
	}
	called := false
	err = sys.Mkprogram(aptpath, func(ctx context.Context, pc *fakesystem.ProgramContext) int {
		if len(pc.Args) != 2 || pc.Args[1] != "update" {
			fmt.Fprintf(pc.Output, "arguments = %v; want [update]\n", pc.Args[1:])
			return 1
		}
		called = true
		return 0
	})
	if err != nil {
		t.Fatal("Mkprogram:", err)
	}

	app := &Applier{
		System: sys,
		Log:    testLogger{t: t},
	}
	err = app.Apply(ctx, cat)
	if err != nil {
		t.Error("Apply:", err)
	}

	if !called {
		t.Error("program not executed")
	}
}

func TestExecIfDepsChanged(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	binpath := filepath.Join(fakesystem.Root, "bin")
	fpath := filepath.Join(fakesystem.Root, "config")
	aptpath := filepath.Join(binpath, "apt-get")
	cat, err := (&catalogStruct{
		Resources: []resource{
			{
				ID:      100,
				Comment: "file",
				Which:   catalog.Resource_Which_file,
				File:    newPlainFile(fpath, []byte("Hello")),
			},
			{
				ID:      42,
				Comment: "apt-get update",
				Deps:    []uint64{100},
				Which:   catalog.Resource_Which_exec,
				Exec: &exec{
					Command: &command{
						Which: catalog.Exec_Command_Which_argv,
						Argv:  []string{aptpath, "update"},
					},
					Condition: execCondition{
						Which:         catalog.Exec_condition_Which_ifDepsChanged,
						IfDepsChanged: []uint64{100},
					},
				},
			},
		},
	}).toCapnp()
	if err != nil {
		t.Fatal("catalogStruct.toCapnp():", err)
	}
	t.Run("trigger", func(t *testing.T) {
		sys := new(fakesystem.System)
		if err := sys.Mkdir(ctx, binpath, 0777); err != nil {
			t.Fatalf("mkdir %s: %v", binpath, err)
		}
		called := false
		err = sys.Mkprogram(aptpath, func(ctx context.Context, pc *fakesystem.ProgramContext) int {
			called = true
			return 0
		})
		if err != nil {
			t.Fatal("Mkprogram:", err)
		}
		app := &Applier{
			System: sys,
			Log:    testLogger{t: t},
		}
		err = app.Apply(ctx, cat)
		if err != nil {
			t.Error("Apply:", err)
		}
		if !called {
			t.Error("program not executed")
		}
	})
	t.Run("no-op", func(t *testing.T) {
		sys := new(fakesystem.System)
		if err := sys.Mkdir(ctx, binpath, 0777); err != nil {
			t.Fatalf("mkdir %s: %v", binpath, err)
		}
		called := false
		err = sys.Mkprogram(aptpath, func(ctx context.Context, pc *fakesystem.ProgramContext) int {
			called = true
			return 0
		})
		if err != nil {
			t.Fatal("Mkprogram:", err)
		}
		if err := system.WriteFile(ctx, sys, fpath, []byte("Hello"), 0666); err != nil {
			t.Fatal("WriteFile:", err)
		}
		app := &Applier{
			System: sys,
			Log:    testLogger{t: t},
		}
		err = app.Apply(ctx, cat)
		if err != nil {
			t.Error("Apply:", err)
		}
		if called {
			t.Error("program executed even though file existed")
		}
	})
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
	Exec  *exec
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

func newSymlinkFile(oldname, newname string) *file {
	f := &file{Path: newname, Which: catalog.File_Which_symlink}
	f.Symlink.Target = oldname
	return f
}

type exec struct {
	Command   *command
	Condition execCondition
}

type execCondition struct {
	Which         catalog.Exec_condition_Which
	OnlyIf        *command
	Unless        *command
	FileAbsent    string
	IfDepsChanged []uint64
}

type command struct {
	Which catalog.Exec_Command_Which
	Argv  []string
	Bash  string

	Env []envVar `capnp:"environment"`
	Dir string   `capnp:"workingDirectory"`
}

type envVar struct {
	Name, Value string
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
