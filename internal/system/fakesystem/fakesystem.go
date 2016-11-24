// Package fakesystem provides an in-memory implementation of the
// interfaces in the system package.
package fakesystem

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/zombiezen/mcm/internal/system"
)

var epoch = time.Date(2016, time.November, 24, 0, 0, 0, 0, time.UTC)

// System is an in-memory implementation of FS and Runner.
// It uses path/filepath for path manipulation.  It is safe to use from
// multiple goroutines.  The zero value is an empty filesystem.
type System struct {
	mu   sync.Mutex
	fs   map[string]*entry
	time time.Time
}

// Program is a function to call when an executable file is run.
type Program func(ctx context.Context, pc *ProgramContext) int

type ProgramContext struct {
	Args   []string
	Env    []string
	Dir    string
	Output io.Writer
}

type entry struct {
	mode    os.FileMode
	modTime time.Time
	content []byte
	program Program
}

func (sys *System) init() {
	if sys.fs != nil {
		return
	}
	sys.time = epoch
	sys.fs = make(map[string]*entry)
	sys.fs["/"] = &entry{
		mode:    os.ModeDir | 0777,
		modTime: sys.time,
	}
}

func (sys *System) stepTime() {
	sys.time = sys.time.Add(1 * time.Second)
}

func (sys *System) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	cleanPath, err := cleanPath("lstat", path)
	if err != nil {
		return nil, err
	}

	defer sys.mu.Unlock()
	defer sys.stepTime()
	sys.mu.Lock()
	sys.init()
	ent := sys.fs[cleanPath]
	if ent == nil {
		return nil, &os.PathError{Op: "lstat", Path: path, Err: os.ErrNotExist}
	}
	return &stat{
		name:    filepath.Base(cleanPath),
		mode:    ent.mode,
		modTime: ent.modTime,
		size:    len(ent.content),
	}, nil
}

func (sys *System) checkParentWritable(op string, path string) error {
	par := sys.fs[filepath.Dir(path)]
	if par == nil {
		return &os.PathError{Op: op, Path: path, Err: os.ErrNotExist}
	}
	if !par.mode.IsDir() {
		return &os.PathError{Op: op, Path: path, Err: errors.New("fake OS: not a directory")}
	}
	if par.mode&0222 == 0 {
		return &os.PathError{Op: op, Path: path, Err: os.ErrPermission}
	}
	return nil
}

func (sys *System) CreateFile(ctx context.Context, path string, mode os.FileMode) (system.FileWriter, error) {
	cleanPath, err := cleanPath("open", path)
	if err != nil {
		return nil, err
	}

	defer sys.mu.Unlock()
	defer sys.stepTime()
	sys.mu.Lock()
	sys.init()
	if sys.fs[cleanPath] != nil {
		return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrExist}
	}
	if err := sys.checkParentWritable("open", path); err != nil {
		return nil, err
	}
	ent := &entry{
		mode:    mode & os.ModePerm,
		modTime: sys.time,
	}
	sys.fs[cleanPath] = ent
	return &openFile{
		mu:  &sys.mu,
		ent: ent,
	}, nil
}

func (sys *System) OpenFile(ctx context.Context, path string) (system.File, error) {
	cleanPath, err := cleanPath("open", path)
	if err != nil {
		return nil, err
	}

	defer sys.mu.Unlock()
	defer sys.stepTime()
	sys.mu.Lock()
	sys.init()
	ent := sys.fs[cleanPath]
	if ent == nil {
		return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
	}
	if !ent.mode.IsRegular() {
		return nil, &os.PathError{Op: "open", Path: path, Err: errors.New("fake OS: not a file")}
	}
	ent.modTime = sys.time
	return &openFile{
		mu:   &sys.mu,
		ent:  ent,
		data: append([]byte(nil), ent.content...),
	}, nil
}

func (sys *System) Mkdir(ctx context.Context, path string, mode os.FileMode) error {
	cleanPath, err := cleanPath("mkdir", path)
	if err != nil {
		return err
	}

	defer sys.mu.Unlock()
	defer sys.stepTime()
	sys.mu.Lock()
	sys.init()
	if sys.fs[cleanPath] != nil {
		return &os.PathError{Op: "mkdir", Path: path, Err: os.ErrExist}
	}
	if err := sys.checkParentWritable("mkdir", path); err != nil {
		return err
	}
	sys.fs[cleanPath] = &entry{
		mode:    os.ModeDir | mode&os.ModePerm,
		modTime: sys.time,
	}
	return nil
}

func (sys *System) Symlink(ctx context.Context, oldname, newname string) error {
	return errors.New("fake system: symlink TODO")
}

func (sys *System) readdir(path string) []string {
	var names []string
	for p := range sys.fs {
		if p == path {
			continue
		}
		dir, base := filepath.Split(p)
		if filepath.Clean(dir) == path {
			names = append(names, base)
		}
	}
	return names
}

func (sys *System) Remove(ctx context.Context, path string) error {
	cleanPath, err := cleanPath("remove", path)
	if err != nil {
		return err
	}

	defer sys.mu.Unlock()
	defer sys.stepTime()
	sys.mu.Lock()
	sys.init()
	ent := sys.fs[cleanPath]
	if ent == nil {
		return &os.PathError{Op: "remove", Path: path, Err: os.ErrNotExist}
	}
	if err := sys.checkParentWritable("remove", path); err != nil {
		return err
	}
	if ent.mode.IsDir() && len(sys.readdir(cleanPath)) > 0 {
		return &os.PathError{Op: "remove", Path: path, Err: errors.New("fake OS: directory not empty")}
	}
	delete(sys.fs, cleanPath)
	return nil
}

// Mkprogram creates a filesystem entry that calls a program when run.
func (sys *System) Mkprogram(path string, prog Program) error {
	cleanPath, err := cleanPath("mkprogram", path)
	if err != nil {
		return err
	}

	defer sys.mu.Unlock()
	defer sys.stepTime()
	sys.mu.Lock()
	sys.init()
	if sys.fs[cleanPath] != nil {
		return &os.PathError{Op: "mkprogram", Path: path, Err: os.ErrExist}
	}
	if err := sys.checkParentWritable("mkprogram", path); err != nil {
		return err
	}
	sys.fs[cleanPath] = &entry{
		mode:    0777,
		modTime: sys.time,
		program: prog,
	}
	return nil
}

func (sys *System) Run(ctx context.Context, cmd *system.Cmd) (output []byte, err error) {
	cleanPath, err := cleanPath("exec", cmd.Path)
	if err != nil {
		return nil, err
	}

	defer sys.mu.Unlock()
	defer sys.stepTime()
	sys.mu.Lock()
	sys.init()
	ent := sys.fs[cleanPath]
	if ent == nil {
		return nil, &os.PathError{Op: "exec", Path: cmd.Path, Err: os.ErrNotExist}
	}
	if ent.mode&0111 == 0 {
		return nil, &os.PathError{Op: "exec", Path: cmd.Path, Err: os.ErrPermission}
	}
	if ent.program == nil {
		return nil, &os.PathError{Op: "exec", Path: cmd.Path, Err: errors.New("fake system: not a program")}
	}
	out := new(bytes.Buffer)
	exit := ent.program(ctx, &ProgramContext{
		Args:   cmd.Args,
		Env:    cmd.Env,
		Dir:    cmd.Dir,
		Output: out,
	})
	if exit != 0 {
		return out.Bytes(), new(exec.ExitError)
	}
	return out.Bytes(), nil
}

var (
	_ system.FS     = (*System)(nil)
	_ system.Runner = (*System)(nil)
)

func cleanPath(op string, path string) (string, error) {
	if !filepath.IsAbs(path) {
		return "", &os.PathError{Op: op, Path: path, Err: errors.New("fake OS: path is not absolute")}
	}
	return filepath.Clean(path), nil
}

type openFile struct {
	mu     *sync.Mutex // pointer to System.mu
	ent    *entry
	data   []byte
	pos    int
	closed bool
}

func (f *openFile) Read(p []byte) (n int, err error) {
	if f.closed {
		return 0, errClosed
	}
	n = copy(p, f.data[f.pos:])
	f.pos += n
	if n >= len(f.data) {
		err = io.EOF
	}
	return
}

func (f *openFile) Write(p []byte) (n int, err error) {
	if f.closed {
		return 0, errClosed
	}
	n = len(p)
	if f.pos < len(f.data) {
		nn := copy(f.data[f.pos:], p)
		p = p[nn:]
		f.pos += nn
		if len(p) == 0 {
			return
		}
	}
	f.data = append(f.data, p...)
	f.pos = len(f.data)
	return
}

func (f *openFile) Seek(offset int64, whence int) (int64, error) {
	if f.closed {
		return 0, errClosed
	}
	old := f.pos
	switch whence {
	case io.SeekCurrent:
		f.pos = int(int64(f.pos) + offset)
	case io.SeekStart:
		f.pos = int(offset)
	case io.SeekEnd:
		f.pos = int(int64(len(f.data)) + offset)
	default:
		return int64(f.pos), fmt.Errorf("fake file: invalid whence %d", whence)
	}
	if f.pos < 0 || f.pos > len(f.data) {
		f.pos = old
		return int64(f.pos), errors.New("fake file: seek past boundaries")
	}
	return int64(f.pos), nil
}

func (f *openFile) Truncate(size int64) error {
	if f.closed {
		return errClosed
	}
	f.data = f.data[:size]
	return nil
}

func (f *openFile) Close() error {
	if f.closed {
		return errClosed
	}
	defer f.mu.Unlock()
	f.mu.Lock()
	f.ent.content = f.data
	f.ent.program = nil
	f.closed = true
	return nil
}

type stat struct {
	name    string
	mode    os.FileMode
	modTime time.Time
	size    int
}

func (s *stat) Name() string       { return s.name }
func (s *stat) Size() int64        { return int64(s.size) }
func (s *stat) Mode() os.FileMode  { return s.mode }
func (s *stat) ModTime() time.Time { return s.modTime }
func (s *stat) IsDir() bool        { return s.mode.IsDir() }
func (s *stat) Sys() interface{}   { return nil }

var errClosed = errors.New("fake file: closed")
