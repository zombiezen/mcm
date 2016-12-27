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

// Package applytests provides an integration test suite for testing algorithms
// that apply catalogs.
package applytests

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/zombiezen/mcm/catalog"
	"github.com/zombiezen/mcm/internal/catpogs"
	"github.com/zombiezen/mcm/internal/system"
)

// A Logger provides a mechanism to print debugging output.
type Logger interface {
	Logf(string, ...interface{})
}

// A FixtureFunc creates a fixture.
type FixtureFunc func(ctx context.Context, log Logger, name string) (Fixture, error)

// Fixture represents the environment for a running integration tests.
// It can apply catalogs to a system and provides some minimal information about
// locations of programs and files necessary for testing.
type Fixture interface {
	Apply(context.Context, catalog.Catalog) error
	System() system.System
	SystemInfo() *SystemInfo
	Close() error
}

// SystemInfo is a list of locations used for tests.
type SystemInfo struct {
	// Root is a writable directory used during the test.
	// It is assumed to be empty when the fixture is created.
	Root string

	// TruePath is a path to a program that always returns success.
	TruePath string
	// FalsePath is a path to a program that always returns failure.
	FalsePath string
	// TouchPath is a path to a program that creates the file named by its single
	// argument.
	TouchPath string
}

// Run runs the test suite as subtests of t.
func Run(t *testing.T, ff FixtureFunc) {
	t.Run("Empty", func(t *testing.T) { emptyTest(t, ff) })
	t.Run("File", func(t *testing.T) { fileTest(t, ff) })
	t.Run("Noop", func(t *testing.T) { noopTest(t, ff) })
	t.Run("NoContentFile", func(t *testing.T) { noContentFileTest(t, ff) })
	t.Run("Link", func(t *testing.T) { linkTest(t, ff) })
	t.Run("Relink", func(t *testing.T) { relinkTest(t, ff) })
	t.Run("SkipFail", func(t *testing.T) { skipFailTest(t, ff) })
	t.Run("Exec", func(t *testing.T) { execTest(t, ff) })
	t.Run("ExecOnlyIf", func(t *testing.T) { execOnlyIfTest(t, ff) })
	t.Run("ExecUnless", func(t *testing.T) { execUnlessTest(t, ff) })
	t.Run("ExecIfDepsChanged", func(t *testing.T) { execIfDepsChangedTest(t, ff) })
}

func startTest(t *testing.T, ff FixtureFunc, name string) (ctx context.Context, f Fixture, done func()) {
	ctx, cancel := context.WithCancel(context.Background())
	f, err := ff(ctx, t, name)
	if err != nil {
		cancel()
		t.Fatal("fixture:", err)
	}
	return ctx, f, func() {
		cancel()
		if err := f.Close(); err != nil {
			t.Error("fixture close:", err)
		}
	}
}

func emptyTest(t *testing.T, ff FixtureFunc) {
	ctx, f, done := startTest(t, ff, "empty")
	defer done()

	c, err := new(catpogs.Catalog).ToCapnp()
	if err != nil {
		t.Fatalf("build empty catalog: %v", err)
	}
	if err = f.Apply(ctx, c); err != nil {
		t.Errorf("run catalog: %v", err)
	}
}

func fileTest(t *testing.T, ff FixtureFunc) {
	ctx, f, done := startTest(t, ff, "file")
	defer done()

	info := f.SystemInfo()
	fpath := filepath.Join(info.Root, "foo.txt")
	const fileContent = "Hello!\n"
	c, err := (&catpogs.Catalog{
		Resources: []*catpogs.Resource{
			{
				ID:      42,
				Comment: "file",
				Which:   catalog.Resource_Which_file,
				File:    catpogs.PlainFile(fpath, []byte(fileContent)),
			},
		},
	}).ToCapnp()
	if err != nil {
		t.Fatalf("build catalog: %v", err)
	}
	err = f.Apply(ctx, c)
	if err != nil {
		t.Errorf("run catalog: %v", err)
	}
	gotContent, err := system.ReadFile(ctx, f.System(), fpath)
	if err != nil {
		t.Errorf("read %s: %v", fpath, err)
	}
	if !bytes.Equal(gotContent, []byte(fileContent)) {
		t.Errorf("content of %s = %q; want %q", fpath, gotContent, fileContent)
	}
}

func noopTest(t *testing.T, ff FixtureFunc) {
	t.Run("NoDeps", func(t *testing.T) {
		ctx, f, done := startTest(t, ff, "noopNoDeps")
		defer done()
		c, err := (&catpogs.Catalog{
			Resources: []*catpogs.Resource{
				{
					ID:      42,
					Comment: "noop",
					Which:   catalog.Resource_Which_noop,
				},
			},
		}).ToCapnp()
		if err != nil {
			t.Fatalf("build catalog: %v", err)
		}
		err = f.Apply(ctx, c)
		if err != nil {
			t.Errorf("run catalog: %v", err)
		}
	})
	t.Run("WithDeps", func(t *testing.T) {
		ctx, f, done := startTest(t, ff, "noopWithDeps")
		defer done()
		root := f.SystemInfo().Root
		fpath1 := filepath.Join(root, "foo1.txt")
		fpath2 := filepath.Join(root, "foo2.txt")
		c, err := (&catpogs.Catalog{
			Resources: []*catpogs.Resource{
				{
					ID:      101,
					Comment: "file1",
					Which:   catalog.Resource_Which_file,
					File:    catpogs.PlainFile(fpath1, []byte("Hello!\n")),
				},
				{
					ID:      102,
					Comment: "file2",
					Which:   catalog.Resource_Which_file,
					File:    catpogs.PlainFile(fpath2, []byte("Hello!\n")),
				},
				{
					ID:      42,
					Comment: "noop",
					Deps:    []uint64{101, 102},
					Which:   catalog.Resource_Which_noop,
				},
			},
		}).ToCapnp()
		if err != nil {
			t.Fatalf("build catalog: %v", err)
		}
		err = f.Apply(ctx, c)
		if err != nil {
			t.Errorf("run catalog: %v", err)
		}
	})
	t.Run("WithDepsChange", func(t *testing.T) {
		ctx, f, done := startTest(t, ff, "noopWithDepsChange")
		defer done()
		root := f.SystemInfo().Root
		fpath1 := filepath.Join(root, "foo1.txt")
		fpath2 := filepath.Join(root, "foo2.txt")
		outpath := filepath.Join(root, "canary.txt")
		c, err := (&catpogs.Catalog{
			Resources: []*catpogs.Resource{
				{
					ID:      101,
					Comment: "file1",
					Which:   catalog.Resource_Which_file,
					File:    catpogs.PlainFile(fpath1, []byte("Hello!\n")),
				},
				{
					ID:      102,
					Comment: "file2",
					Which:   catalog.Resource_Which_file,
					File:    catpogs.PlainFile(fpath2, []byte("Hello!\n")),
				},
				{
					ID:      42,
					Comment: "noop",
					Deps:    []uint64{101, 102},
					Which:   catalog.Resource_Which_noop,
				},
				{
					ID:      200,
					Comment: "canary",
					Deps:    []uint64{42},
					Which:   catalog.Resource_Which_exec,
					Exec: &catpogs.Exec{
						Command: &catpogs.Command{
							Which: catalog.Exec_Command_Which_argv,
							Argv:  []string{f.SystemInfo().TouchPath, outpath},
						},
						Condition: catpogs.ExecCondition{
							Which:         catalog.Exec_condition_Which_ifDepsChanged,
							IfDepsChanged: []uint64{42},
						},
					},
				},
			},
		}).ToCapnp()
		if err != nil {
			t.Fatalf("build catalog: %v", err)
		}
		err = f.Apply(ctx, c)
		if err != nil {
			t.Errorf("run catalog: %v", err)
		}
		ok, err := fileExists(ctx, f.System(), outpath)
		if err != nil {
			t.Error("fileExists:", err)
		} else if !ok {
			t.Errorf("%q does not exist; noop resource did not report changed", outpath)
		}
	})
}

func noContentFileTest(t *testing.T, ff FixtureFunc) {
	t.Run("Exists", func(t *testing.T) {
		ctx, f, done := startTest(t, ff, "noContentExists")
		defer done()
		fpath := filepath.Join(f.SystemInfo().Root, "foo.txt")
		if err := system.WriteFile(ctx, f.System(), fpath, []byte("data\n"), 0666); err != nil {
			t.Fatal("WriteFile:", err)
		}
		c, err := (&catpogs.Catalog{
			Resources: []*catpogs.Resource{
				{
					ID:      42,
					Comment: "file",
					Which:   catalog.Resource_Which_file,
					File:    catpogs.PlainFile(fpath, nil),
				},
			},
		}).ToCapnp()
		if err != nil {
			t.Fatalf("build catalog: %v", err)
		}
		err = f.Apply(ctx, c)
		if err != nil {
			t.Errorf("run catalog: %v", err)
		}
	})
	t.Run("NotExists", func(t *testing.T) {
		ctx, f, done := startTest(t, ff, "noContentNotExists")
		defer done()
		fpath := filepath.Join(f.SystemInfo().Root, "foo.txt")
		c, err := (&catpogs.Catalog{
			Resources: []*catpogs.Resource{
				{
					ID:      42,
					Comment: "file",
					Which:   catalog.Resource_Which_file,
					File:    catpogs.PlainFile(fpath, nil),
				},
			},
		}).ToCapnp()
		if err != nil {
			t.Fatalf("build catalog: %v", err)
		}
		err = f.Apply(ctx, c)
		if err == nil {
			t.Error("run catalog did not fail as expected")
		}
		if exists, err := fileExists(ctx, f.System(), fpath); err != nil {
			t.Error("fileExists:", err)
		} else if exists {
			t.Errorf("file %q exists; applier should not have created", fpath)
		}
	})
}

func linkTest(t *testing.T, ff FixtureFunc) {
	ctx, f, done := startTest(t, ff, "link")
	defer done()
	root := f.SystemInfo().Root
	fpath := filepath.Join(root, "foo")
	lpath := filepath.Join(root, "link")
	c, err := (&catpogs.Catalog{
		Resources: []*catpogs.Resource{
			{
				ID:      42,
				Comment: "file",
				Which:   catalog.Resource_Which_file,
				File:    catpogs.PlainFile(fpath, []byte("Hello")),
			},
			{
				ID:      100,
				Deps:    []uint64{42},
				Comment: "link",
				Which:   catalog.Resource_Which_file,
				File:    catpogs.SymlinkFile(fpath, lpath),
			},
		},
	}).ToCapnp()
	if err != nil {
		t.Fatalf("build catalog: %v", err)
	}
	err = f.Apply(ctx, c)
	if err != nil {
		t.Errorf("run catalog: %v", err)
	}

	sys := f.System()
	if info, err := sys.Lstat(ctx, lpath); err == nil {
		if info.Mode()&os.ModeType != os.ModeSymlink {
			t.Errorf("Lstat(%q).Mode() = %v; want symlink", lpath, info.Mode())
		}
	} else {
		t.Errorf("Lstat(%q): %v", lpath, err)
	}
	if target, err := sys.Readlink(ctx, lpath); err == nil {
		if target != fpath {
			t.Errorf("Readlink(%q) = %q; want %q", lpath, target, fpath)
		}
	} else {
		t.Errorf("Readlink(%q): %v", lpath, err)
	}
}

func relinkTest(t *testing.T, ff FixtureFunc) {
	ctx, f, done := startTest(t, ff, "relink")
	defer done()
	root := f.SystemInfo().Root
	f1path := filepath.Join(root, "foo")
	f2path := filepath.Join(root, "bar")
	lpath := filepath.Join(root, "link")
	c, err := (&catpogs.Catalog{
		Resources: []*catpogs.Resource{
			{
				ID:      42,
				Comment: "link",
				Which:   catalog.Resource_Which_file,
				File:    catpogs.SymlinkFile(f2path, lpath),
			},
		},
	}).ToCapnp()
	if err != nil {
		t.Fatalf("build catalog: %v", err)
	}
	sys := f.System()
	if err := system.WriteFile(ctx, sys, f1path, []byte("File 1"), 0666); err != nil {
		t.Fatal("WriteFile 1:", err)
	}
	if err := system.WriteFile(ctx, sys, f2path, []byte("File 2"), 0666); err != nil {
		t.Fatal("WriteFile 2:", err)
	}
	if err := sys.Symlink(ctx, f1path, lpath); err != nil {
		t.Fatalf("os.Symlink %s -> %s: %v", lpath, f1path, err)
	}
	err = f.Apply(ctx, c)
	if err != nil {
		t.Errorf("run catalog: %v", err)
	}

	if info, err := sys.Lstat(ctx, lpath); err == nil {
		if info.Mode()&os.ModeType != os.ModeSymlink {
			t.Errorf("Lstat(%q).Mode() = %v; want symlink", lpath, info.Mode())
		}
	} else {
		t.Errorf("Lstat(%q): %v", lpath, err)
	}
	if target, err := sys.Readlink(ctx, lpath); err == nil {
		if target != f2path {
			t.Errorf("Readlink(%q) = %q; want %q", lpath, target, f2path)
		}
	} else {
		t.Errorf("Readlink(%q): %v", lpath, err)
	}
}

func skipFailTest(t *testing.T, ff FixtureFunc) {
	ctx, f, done := startTest(t, ff, "skipFail")
	defer done()
	root := f.SystemInfo().Root
	f1path := filepath.Join(root, "foo")
	f2path := filepath.Join(root, "bar")
	f3path := filepath.Join(root, "baz")
	canaryPath := filepath.Join(root, "canary")
	c, err := (&catpogs.Catalog{
		Resources: []*catpogs.Resource{
			{
				ID:      101,
				Comment: "file 1",
				Which:   catalog.Resource_Which_file,
				File:    catpogs.PlainFile(f1path, []byte("foo")),
			},
			{
				ID:      102,
				Deps:    []uint64{101},
				Comment: "file 2",
				Which:   catalog.Resource_Which_file,
				File:    catpogs.PlainFile(f2path, []byte("bar")),
			},
			{
				ID:      103,
				Deps:    []uint64{102},
				Comment: "file 3",
				Which:   catalog.Resource_Which_file,
				File:    catpogs.PlainFile(f3path, []byte("baz")),
			},
			{
				ID:      200,
				Comment: "canary file - not dependent on other files",
				Which:   catalog.Resource_Which_file,
				File:    catpogs.PlainFile(canaryPath, []byte("tweet!")),
			},
		},
	}).ToCapnp()
	if err != nil {
		t.Fatalf("build catalog: %v", err)
	}
	// Create a directory to cause file 1 to fail.
	sys := f.System()
	if err := sys.Mkdir(ctx, f1path, 0777); err != nil {
		t.Fatal("mkdir:", err)
	}
	err = f.Apply(ctx, c)
	t.Logf("run catalog: %v", err)
	if err == nil {
		t.Error("run catalog did not return an error")
	}

	if exists, err := fileExists(ctx, sys, f2path); exists || err != nil {
		t.Errorf("fileExists(%q) = %t, %v; false, nil", f2path, exists, err)
	}
	if exists, err := fileExists(ctx, sys, f3path); exists || err != nil {
		t.Errorf("fileExists(%q) = %t, %v; false, nil", f3path, exists, err)
	}
	if exists, err := fileExists(ctx, sys, canaryPath); !exists || err != nil {
		t.Errorf("fileExists(%q) = %t, %v; false, nil", canaryPath, exists, err)
	}
}

func execTest(t *testing.T, ff FixtureFunc) {
	ctx, f, done := startTest(t, ff, "exec")
	defer done()
	info := f.SystemInfo()
	fpath := filepath.Join(info.Root, "canary")
	c, err := (&catpogs.Catalog{
		Resources: []*catpogs.Resource{
			{
				ID:      42,
				Comment: "exec",
				Which:   catalog.Resource_Which_exec,
				Exec: &catpogs.Exec{
					Command: &catpogs.Command{
						Which: catalog.Exec_Command_Which_argv,
						Argv:  []string{info.TouchPath, fpath},
					},
				},
			},
		},
	}).ToCapnp()
	if err != nil {
		t.Fatalf("build catalog: %v", err)
	}
	err = f.Apply(ctx, c)
	if err != nil {
		t.Errorf("run catalog: %v", err)
	}
	if exists, err := fileExists(ctx, f.System(), fpath); err != nil {
		t.Error("fileExists:", err)
	} else if !exists {
		t.Errorf("file %q not created", fpath)
	}
}

func execOnlyIfTest(t *testing.T, ff FixtureFunc) {
	run := func(t *testing.T, name string, cond bool, want bool) {
		ctx, f, done := startTest(t, ff, name)
		defer done()
		info := f.SystemInfo()
		fpath := filepath.Join(info.Root, "canary")
		var condExe string
		if cond {
			condExe = info.TruePath
		} else {
			condExe = info.FalsePath
		}
		c, err := (&catpogs.Catalog{
			Resources: []*catpogs.Resource{
				{
					ID:      42,
					Comment: "touch canary",
					Which:   catalog.Resource_Which_exec,
					Exec: &catpogs.Exec{
						Command: &catpogs.Command{
							Which: catalog.Exec_Command_Which_argv,
							Argv:  []string{info.TouchPath, fpath},
						},
						Condition: catpogs.ExecCondition{
							Which: catalog.Exec_condition_Which_onlyIf,
							OnlyIf: &catpogs.Command{
								Which: catalog.Exec_Command_Which_argv,
								Argv:  []string{condExe},
							},
						},
					},
				},
			},
		}).ToCapnp()
		if err != nil {
			t.Fatalf("build catalog: %v", err)
		}
		err = f.Apply(ctx, c)
		if err != nil {
			t.Errorf("run catalog: %v", err)
		}
		if exists, err := fileExists(ctx, f.System(), fpath); err != nil {
			t.Error("fileExists:", err)
		} else if exists != want {
			t.Errorf("existence of %q = %t; want %t", fpath, exists, want)
		}
	}
	t.Run("True", func(t *testing.T) {
		run(t, "execOnlyIfTrue", true, true)
	})
	t.Run("False", func(t *testing.T) {
		run(t, "execOnlyIfFalse", false, false)
	})
}

func execUnlessTest(t *testing.T, ff FixtureFunc) {
	run := func(t *testing.T, name string, cond bool, want bool) {
		ctx, f, done := startTest(t, ff, name)
		defer done()
		info := f.SystemInfo()
		fpath := filepath.Join(info.Root, "canary")
		var condExe string
		if cond {
			condExe = info.TruePath
		} else {
			condExe = info.FalsePath
		}
		c, err := (&catpogs.Catalog{
			Resources: []*catpogs.Resource{
				{
					ID:      42,
					Comment: "touch canary",
					Which:   catalog.Resource_Which_exec,
					Exec: &catpogs.Exec{
						Command: &catpogs.Command{
							Which: catalog.Exec_Command_Which_argv,
							Argv:  []string{info.TouchPath, fpath},
						},
						Condition: catpogs.ExecCondition{
							Which: catalog.Exec_condition_Which_unless,
							Unless: &catpogs.Command{
								Which: catalog.Exec_Command_Which_argv,
								Argv:  []string{condExe},
							},
						},
					},
				},
			},
		}).ToCapnp()
		if err != nil {
			t.Fatalf("build catalog: %v", err)
		}
		err = f.Apply(ctx, c)
		if err != nil {
			t.Errorf("run catalog: %v", err)
		}
		if exists, err := fileExists(ctx, f.System(), fpath); err != nil {
			t.Error("fileExists:", err)
		} else if exists != want {
			t.Errorf("existence of %q = %t; want %t", fpath, exists, want)
		}
	}
	t.Run("True", func(t *testing.T) {
		run(t, "execUnlessTrue", true, false)
	})
	t.Run("False", func(t *testing.T) {
		run(t, "execUnlessFalse", false, true)
	})
}

func execIfDepsChangedTest(t *testing.T, ff FixtureFunc) {
	const fileName = "config"
	const fileContent = "Hello"
	const canaryName = "canary"
	makeCatalog := func(info *SystemInfo) (catalog.Catalog, error) {
		fpath := filepath.Join(info.Root, fileName)
		cat, err := (&catpogs.Catalog{
			Resources: []*catpogs.Resource{
				{
					ID:      100,
					Comment: "file",
					Which:   catalog.Resource_Which_file,
					File:    catpogs.PlainFile(fpath, []byte(fileContent)),
				},
				{
					ID:      42,
					Comment: "touch " + canaryName,
					Deps:    []uint64{100},
					Which:   catalog.Resource_Which_exec,
					Exec: &catpogs.Exec{
						Command: &catpogs.Command{
							Which: catalog.Exec_Command_Which_argv,
							Argv:  []string{info.TouchPath, filepath.Join(info.Root, canaryName)},
						},
						Condition: catpogs.ExecCondition{
							Which:         catalog.Exec_condition_Which_ifDepsChanged,
							IfDepsChanged: []uint64{100},
						},
					},
				},
			},
		}).ToCapnp()
		if err != nil {
			return catalog.Catalog{}, fmt.Errorf("build catalog: %v", err)
		}
		return cat, nil
	}
	t.Run("trigger", func(t *testing.T) {
		ctx, f, done := startTest(t, ff, "execIfDepsChanged")
		defer done()
		info := f.SystemInfo()
		c, err := makeCatalog(info)
		if err != nil {
			t.Fatal(err)
		}
		err = f.Apply(ctx, c)
		if err != nil {
			t.Errorf("run catalog: %v", err)
		}
		canaryPath := filepath.Join(info.Root, canaryName)
		if exists, err := fileExists(ctx, f.System(), canaryPath); err != nil {
			t.Errorf("checking for %q existence: %v", canaryPath, err)
		} else if !exists {
			t.Errorf("file %q does not exist; exec not run", canaryPath)
		}
	})
	t.Run("no-op", func(t *testing.T) {
		ctx, f, done := startTest(t, ff, "execIfDepsNotChanged")
		defer done()
		sys := f.System()
		info := f.SystemInfo()
		fpath := filepath.Join(info.Root, fileName)
		if err := system.WriteFile(ctx, sys, fpath, []byte(fileContent), 0666); err != nil {
			t.Fatal("WriteFile:", err)
		}
		c, err := makeCatalog(f.SystemInfo())
		if err != nil {
			t.Fatal(err)
		}
		err = f.Apply(ctx, c)
		if err != nil {
			t.Errorf("run catalog: %v", err)
		}
		canaryPath := filepath.Join(info.Root, canaryName)
		if exists, err := fileExists(ctx, f.System(), canaryPath); err != nil {
			t.Errorf("checking for %q existence: %v", canaryPath, err)
		} else if exists {
			t.Errorf("file %q exists; exec was run, even though deps not changed", canaryPath)
		}
	})
}

func fileExists(ctx context.Context, fs system.FS, path string) (bool, error) {
	_, err := fs.Lstat(ctx, path)
	if system.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
