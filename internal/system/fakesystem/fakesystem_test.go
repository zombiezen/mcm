package fakesystem

import (
	"context"
	"fmt"
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
