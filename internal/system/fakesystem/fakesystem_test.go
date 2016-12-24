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

package fakesystem

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zombiezen/mcm/internal/system"
)

func TestZero(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sys := new(System)
	if info, err := sys.Lstat(ctx, Root); err != nil {
		t.Errorf("sys.Lstat(ctx, %q) = _, %v", Root, err)
	} else if !info.IsDir() || !info.Mode().IsDir() {
		t.Errorf("sys.Lstat(ctx, %q).Mode() = %v, nil; want directory", Root, info.Mode())
	}
}

func TestMkdir(t *testing.T) {
	t.Run("mkdir /foo", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sys := new(System)
		dirpath := filepath.Join(Root, "foo")
		if err := mkdir(ctx, t, sys, dirpath); err != nil {
			t.Error(err)
		}
		if info, err := sys.Lstat(ctx, dirpath); err != nil {
			t.Errorf("sys.Lstat(ctx, %q) = _, %v; want nil", dirpath, err)
		} else if !info.IsDir() {
			t.Errorf("sys.Lstat(ctx, %q).Mode() = %v; want directory", dirpath, info.Mode())
		}
	})
	t.Run("mkdir /foo; mkdir /foo/bar", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sys := new(System)
		dirpath1 := filepath.Join(Root, "foo")
		if err := mkdir(ctx, t, sys, dirpath1); err != nil {
			t.Error(err)
		}
		dirpath2 := filepath.Join(dirpath1, "bar")
		if err := mkdir(ctx, t, sys, dirpath2); err != nil {
			t.Error(err)
		}
		if info, err := sys.Lstat(ctx, dirpath2); err != nil {
			t.Errorf("sys.Lstat(ctx, %q) = _, %v; want nil", dirpath2, err)
		} else if !info.IsDir() {
			t.Errorf("sys.Lstat(ctx, %q).Mode() = %v; want directory", dirpath2, info.Mode())
		}
	})
}

func TestRemove(t *testing.T) {
	emptyDirPath := filepath.Join(Root, "emptydir")
	filePath := filepath.Join(Root, "file")
	filledDirPath := filepath.Join(Root, "nonemptydir")
	dirFilePath := filepath.Join(filledDirPath, "baz")
	fileLinkPath := filepath.Join(Root, "link")
	newSystem := func(ctx context.Context, log logger) (*System, error) {
		sys := new(System)
		if err := mkdir(ctx, log, sys, emptyDirPath); err != nil {
			return nil, err
		}
		if err := mkfile(ctx, log, sys, filePath, []byte("Hello")); err != nil {
			return nil, err
		}
		if err := mkdir(ctx, log, sys, filledDirPath); err != nil {
			return nil, err
		}
		if err := mkfile(ctx, log, sys, dirFilePath, []byte("Goodbye")); err != nil {
			return nil, err
		}
		if err := mklink(ctx, log, sys, dirFilePath, fileLinkPath); err != nil {
			return nil, err
		}
		return sys, nil
	}

	tests := []struct {
		path       string
		fails      bool
		isNotExist bool
	}{
		{path: emptyDirPath},
		{path: filePath},
		{path: filledDirPath, fails: true},
		{path: filepath.Join(Root, "nonexistent"), fails: true, isNotExist: true},
		{path: Root, fails: true},
		{path: fileLinkPath},
	}
	for i := range tests {
		test := tests[i]
		t.Run("\""+test.path+"\"", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			sys, err := newSystem(ctx, t)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("sys.Remove(ctx, %q)", test.path)
			err = sys.Remove(ctx, test.path)
			if !test.fails {
				if err != nil {
					t.Errorf("sys.Remove(ctx, %q) = %v; want nil", test.path, err)
				}
				if _, err := sys.Lstat(ctx, test.path); !system.IsNotExist(err) {
					t.Errorf("sys.Lstat(ctx, %q) = _, %v; want is not exist", test.path, err)
				}
			} else {
				if err == nil {
					t.Errorf("sys.Remove(ctx, %q) = nil; want non-nil", test.path)
				} else if test.isNotExist && !system.IsNotExist(err) {
					t.Errorf("sys.Remove(ctx, %q) = %v; want not exist", test.path, err)
				}
				if !test.isNotExist {
					if _, err := sys.Lstat(ctx, test.path); err != nil {
						t.Errorf("sys.Lstat(ctx, %q) = _, %v; want nil", test.path, err)
					}
				}
			}
		})
	}
}

func TestSymlink(t *testing.T) {
	dpath := filepath.Join(Root, "dir")
	fpath := filepath.Join(dpath, "foo.txt")
	const fileContent = "Hello"
	newSystem := func(ctx context.Context, log logger) (*System, error) {
		sys := new(System)
		if err := mkdir(ctx, log, sys, dpath); err != nil {
			return nil, err
		}
		if err := mkfile(ctx, log, sys, fpath, []byte(fileContent)); err != nil {
			return nil, err
		}
		return sys, nil
	}
	t.Run("file", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sys, err := newSystem(ctx, t)
		if err != nil {
			t.Fatal(err)
		}

		lpath := filepath.Join(Root, "mylink")
		t.Logf("sys.Symlink(ctx, %q, %q)", fpath, lpath)
		if err := sys.Symlink(ctx, fpath, lpath); err != nil {
			t.Errorf("sys.Symlink(ctx, %q, %q): %v", fpath, lpath, err)
		}

		if target, err := sys.Readlink(ctx, lpath); err != nil {
			t.Errorf("sys.Readlink(ctx, %q): %v", lpath, err)
		} else if target != fpath {
			t.Errorf("sys.Readlink(ctx, %q) = %q; want %q", lpath, target, fpath)
		}
		f, err := sys.OpenFile(ctx, lpath)
		if err != nil {
			t.Fatalf("sys.OpenFile(ctx, %q): %v", lpath, err)
		}
		defer f.Close()
		content, err := ioutil.ReadAll(f)
		if err != nil {
			t.Errorf("ioutil.ReadAll(sys.OpenFile(ctx, %q)): %v", lpath, err)
		}
		if !bytes.Equal(content, []byte(fileContent)) {
			t.Errorf("ioutil.ReadAll(sys.OpenFile(ctx, %q)) = %q; want %q", lpath, content, fileContent)
		}
	})
	t.Run("directory", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sys, err := newSystem(ctx, t)
		if err != nil {
			t.Fatal(err)
		}

		lpath := filepath.Join(Root, "mylink")
		t.Logf("sys.Symlink(ctx, %q, %q)", dpath, lpath)
		if err := sys.Symlink(ctx, dpath, lpath); err != nil {
			t.Errorf("sys.Symlink(ctx, %q, %q): %v", dpath, lpath, err)
		}

		if target, err := sys.Readlink(ctx, lpath); err != nil {
			t.Errorf("sys.Readlink(ctx, %q): %v", lpath, err)
		} else if target != dpath {
			t.Errorf("sys.Readlink(ctx, %q) = %q; want %q", lpath, target, dpath)
		}
		lfpath := filepath.Join(lpath, filepath.Base(fpath))
		f, err := sys.OpenFile(ctx, lfpath)
		if err != nil {
			t.Fatalf("sys.OpenFile(ctx, %q): %v", lfpath, err)
		}
		defer f.Close()
		content, err := ioutil.ReadAll(f)
		if err != nil {
			t.Errorf("ioutil.ReadAll(sys.OpenFile(ctx, %q)): %v", lfpath, err)
		}
		if !bytes.Equal(content, []byte(fileContent)) {
			t.Errorf("ioutil.ReadAll(sys.OpenFile(ctx, %q)) = %q; want %q", lfpath, content, fileContent)
		}
	})
}

func TestRun(t *testing.T) {
	binDir := filepath.Join(Root, "bin")
	progPath := filepath.Join(binDir, "program")
	newSystem := func(ctx context.Context, log logger, prog Program) (*System, error) {
		sys := new(System)
		if err := mkdir(ctx, log, sys, binDir); err != nil {
			return nil, err
		}
		log.Logf("sys.Mkprogram(%q, ...)", progPath)
		if err := sys.Mkprogram(progPath, prog); err != nil {
			return nil, fmt.Errorf("sys.Mkprogram(%q, ...): %v", progPath, err)
		}
		return sys, nil
	}
	t.Run("called", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		called := false
		sys, err := newSystem(ctx, t, func(ctx context.Context, pc *ProgramContext) int {
			called = true
			return 0
		})
		if err != nil {
			t.Fatal(err)
		}
		t.Log("sys.Run(...)")
		_, err = sys.Run(ctx, &system.Cmd{
			Path: progPath,
			Args: []string{progPath},
			Env:  []string{},
			Dir:  Root,
		})
		if !called {
			t.Error("program function not called")
		}
		if err != nil {
			t.Errorf("sys.Run(...): %v", err)
		}
	})
	t.Run("fail exit", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sys, err := newSystem(ctx, t, func(ctx context.Context, pc *ProgramContext) int {
			return 1
		})
		if err != nil {
			t.Fatal(err)
		}
		t.Log("sys.Run(...)")
		_, err = sys.Run(ctx, &system.Cmd{
			Path: progPath,
			Args: []string{progPath},
			Env:  []string{},
			Dir:  Root,
		})
		if _, ok := err.(*exec.ExitError); !ok {
			t.Errorf("sys.Run(...) = _, %v; want os/exec.ExitError", err)
		}
	})
	t.Run("nil stdin", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sys, err := newSystem(ctx, t, func(ctx context.Context, pc *ProgramContext) int {
			_, err := io.Copy(pc.Output, pc.Input)
			if err != nil {
				return 1
			}
			return 0
		})
		if err != nil {
			t.Fatal(err)
		}
		t.Log("sys.Run(...)")
		out, err := sys.Run(ctx, &system.Cmd{
			Path:  progPath,
			Args:  []string{progPath},
			Env:   []string{},
			Dir:   Root,
			Stdin: nil,
		})
		if err != nil {
			t.Errorf("sys.Run(...): %v", err)
		}
		if len(out) != 0 {
			t.Errorf("sys.Run(...) output = %q; want \"\"", out)
		}
	})
	t.Run("stdin", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sys, err := newSystem(ctx, t, func(ctx context.Context, pc *ProgramContext) int {
			_, err := io.Copy(pc.Output, pc.Input)
			if err != nil {
				return 1
			}
			return 0
		})
		if err != nil {
			t.Fatal(err)
		}
		t.Log("sys.Run(...)")
		const want = "xyzzy"
		out, err := sys.Run(ctx, &system.Cmd{
			Path:  progPath,
			Args:  []string{progPath},
			Env:   []string{},
			Dir:   Root,
			Stdin: strings.NewReader(want),
		})
		if err != nil {
			t.Errorf("sys.Run(...): %v", err)
		}
		if !bytes.Equal(out, []byte(want)) {
			t.Errorf("sys.Run(...) output = %q; want %q", out, want)
		}
	})
}

func TestPathParts(t *testing.T) {
	type testCase struct {
		path  string
		parts []string
	}

	t.Run("unix", func(t *testing.T) {
		if filepath.Separator != '/' {
			t.Skip("not a POSIX system")
		}
		tests := []testCase{
			{"", nil},
			{"foo/bar", nil},
			{"/", []string{"/"}},
			{"/foo/bar", []string{"/", "foo", "bar"}},
		}
		for _, test := range tests {
			parts := pathParts(test.path)
			if !stringSlicesEqual(parts, test.parts) {
				t.Errorf("pathParts(%q) = %q; want %q", test.path, parts, test.parts)
			}
		}
	})
	t.Run("windows", func(t *testing.T) {
		if filepath.Separator != '\\' {
			t.Skip("not a Windows system")
		}
		tests := []testCase{
			{"", nil},
			{`foo\bar`, nil},
			{`C:\`, []string{`C:\`}},
			{`C:\foo\bar`, []string{`C:\`, "foo", "bar"}},
		}
		for _, test := range tests {
			parts := pathParts(test.path)
			if !stringSlicesEqual(parts, test.parts) {
				t.Errorf("pathParts(%q) = %q; want %q", test.path, parts, test.parts)
			}
		}
	})
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

type logger interface {
	Logf(string, ...interface{})
}

func mkdir(ctx context.Context, log logger, fs system.FS, path string) error {
	log.Logf("sys.Mkdir(ctx, %q, 0777)", path)
	if err := fs.Mkdir(ctx, path, 0777); err != nil {
		return fmt.Errorf("sys.Mkdir(ctx, %q, 0777): %v", path, err)
	}
	return nil
}

func mkfile(ctx context.Context, log logger, fs system.FS, path string, content []byte) error {
	log.Logf("system.WriteFile(ctx, sys, %q, %q, 0666)", path, content)
	if err := system.WriteFile(ctx, fs, path, content, 0666); err != nil {
		return fmt.Errorf("system.WriteFile(ctx, sys, %q, %q, 0666): %v", path, content, err)
	}
	return nil
}

func mklink(ctx context.Context, log logger, fs system.FS, oldname, newname string) error {
	log.Logf("system.Symlink(ctx, sys, %q, %q)", oldname, newname)
	if err := fs.Symlink(ctx, oldname, newname); err != nil {
		return fmt.Errorf("sys.Symlink(ctx, %q, %q): %v", oldname, newname, err)
	}
	return nil
}
