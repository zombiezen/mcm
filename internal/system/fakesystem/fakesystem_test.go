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
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
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
