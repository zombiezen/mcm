// Package system provides an interface for interacting with an
// operating system.  It is largely similar to the os package's API
// surface, but presented as an interface for testing purposes (and
// possibly remote execution).
package system

import (
	"context"
	"errors"
	"io"
	"os"
)

// An FS provides access to a filesystem.  Paths are filesystem-specific
// and are generally required to be absolute.  An FS must be safe to
// call from multiple goroutines.
type FS interface {
	Lstat(ctx context.Context, path string) (os.FileInfo, error)
	Mkdir(ctx context.Context, path string, mode os.FileMode) error
	Remove(ctx context.Context, path string) error
	Symlink(ctx context.Context, oldname, newname string) error

	// CreateFile creates the named file, returning an error if it already exists.
	CreateFile(ctx context.Context, path string, mode os.FileMode) (FileWriter, error)

	// OpenFile opens the named file for reading and writing.
	// An error is returned if the file does not already exist.
	OpenFile(ctx context.Context, path string) (File, error)
}

// File represents an open file.
type File interface {
	io.Reader
	io.Writer
	io.Seeker
	Truncate(size int64) error
	io.Closer
}

// FileWriter represents a file open for writing.
type FileWriter interface {
	io.Writer
	io.Closer
}

// A Runner runs processes.  A Runner must be safe to call from
// multiple goroutines.
type Runner interface {
	Run(ctx context.Context, cmd *Cmd) (output []byte, err error)
}

// A Cmd describes a process to execute on a system.
// Its fields behave the same as the corresponding ones in os/exec.Cmd.
type Cmd struct {
	Path string
	Args []string
	Env  []string
	Dir  string
}

func IsExist(err error) bool    { return os.IsExist(err) }
func IsNotExist(err error) bool { return os.IsNotExist(err) }

func WriteFile(ctx context.Context, fs FS, path string, content []byte, mode os.FileMode) error {
	w, err := fs.CreateFile(ctx, path, mode)
	if IsExist(err) {
		f, err := fs.OpenFile(ctx, path)
		if err != nil {
			return err
		}
		if err = f.Truncate(0); err != nil {
			f.Close()
			return err
		}
		w = f
	} else if err != nil {
		return err
	}
	_, err = w.Write(content)
	cerr := w.Close()
	if err != nil {
		return err
	}
	if cerr != nil {
		return cerr
	}
	return nil
}

// Stub returns errors for all of the FS and Runner methods.
type Stub struct{}

func (Stub) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	return nil, &os.PathError{Op: "lstat", Path: path, Err: errNotImplemented}
}

func (Stub) Mkdir(ctx context.Context, path string, mode os.FileMode) error {
	return &os.PathError{Op: "mkdir", Path: path, Err: errNotImplemented}
}

func (Stub) Remove(ctx context.Context, path string) error {
	return &os.PathError{Op: "remove", Path: path, Err: errNotImplemented}
}

func (Stub) Symlink(ctx context.Context, oldname, newname string) error {
	return &os.LinkError{Op: "symlink", Old: oldname, New: newname, Err: errNotImplemented}
}

func (Stub) CreateFile(ctx context.Context, path string, mode os.FileMode) (io.WriteCloser, error) {
	return nil, &os.PathError{Op: "open", Path: path, Err: errNotImplemented}
}

func (Stub) OpenFile(ctx context.Context, path string) (File, error) {
	return nil, &os.PathError{Op: "open", Path: path, Err: errNotImplemented}
}

func (Stub) Run(ctx context.Context, cmd *Cmd) (output []byte, err error) {
	return nil, errNotImplemented
}

var errNotImplemented = errors.New("system stub: not implemented")
