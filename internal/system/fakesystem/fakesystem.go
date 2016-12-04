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

const Root = system.LocalRoot

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
	link    string
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

func (sys *System) resolve(path string) string {
	parts := pathParts(path)
	if len(parts) == 0 {
		return path
	}
	curr, parts := parts[0], parts[1:]
	for i, p := range parts {
		var ok bool
		curr, ok = sys.readlink(filepath.Join(curr, p))
		if !ok {
			return filepath.Join(curr, filepath.Join(parts[i+1:]...))
		}
	}
	return curr
}

func (sys *System) readlink(path string) (string, bool) {
	for {
		ent := sys.fs[path]
		if ent == nil {
			return path, false
		}
		if ent.mode&os.ModeType != os.ModeSymlink {
			return path, true
		}
		if filepath.IsAbs(ent.link) {
			path = ent.link
		} else {
			path = filepath.Join(filepath.Dir(path), ent.link)
		}
	}
}

func (sys *System) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	wrap := pathErrorFunc("lstat", path)
	path, err := cleanPath(path)
	if err != nil {
		return nil, wrap(err)
	}

	defer sys.mu.Unlock()
	defer sys.stepTime()
	sys.mu.Lock()
	sys.init()
	dir, name := filepath.Split(path)
	dir = sys.resolve(dir)
	ent := sys.fs[filepath.Join(dir, name)]
	if ent == nil {
		return nil, wrap(os.ErrNotExist)
	}
	return &stat{
		name:    name,
		mode:    ent.mode,
		modTime: ent.modTime,
		size:    len(ent.content),
	}, nil
}

func (sys *System) mkentry(path string, mode os.FileMode) (*entry, error) {
	dir, name := filepath.Split(path)
	dir = sys.resolve(dir)
	par := sys.fs[dir]
	if par == nil {
		return nil, os.ErrNotExist
	}
	if !par.mode.IsDir() {
		return nil, errors.New("fake OS: not a directory")
	}
	if par.mode&0222 == 0 {
		return nil, os.ErrPermission
	}
	path = filepath.Join(dir, name)
	if sys.fs[path] != nil {
		return nil, os.ErrExist
	}
	ent := &entry{
		mode:    mode,
		modTime: sys.time,
	}
	sys.fs[path] = ent
	return ent, nil
}

func (sys *System) CreateFile(ctx context.Context, path string, mode os.FileMode) (system.FileWriter, error) {
	wrap := pathErrorFunc("open", path)
	path, err := cleanPath(path)
	if err != nil {
		return nil, wrap(err)
	}

	defer sys.mu.Unlock()
	defer sys.stepTime()
	sys.mu.Lock()
	sys.init()
	ent, err := sys.mkentry(path, mode&os.ModePerm)
	if err != nil {
		return nil, wrap(err)
	}
	return &openFile{
		mu:  &sys.mu,
		ent: ent,
	}, nil
}

func (sys *System) OpenFile(ctx context.Context, path string) (system.File, error) {
	wrap := pathErrorFunc("open", path)
	path, err := cleanPath(path)
	if err != nil {
		return nil, wrap(err)
	}

	defer sys.mu.Unlock()
	defer sys.stepTime()
	sys.mu.Lock()
	sys.init()
	ent := sys.fs[sys.resolve(path)]
	if ent == nil {
		return nil, wrap(os.ErrNotExist)
	}
	if !ent.mode.IsRegular() {
		return nil, wrap(errors.New("fake OS: not a file"))
	}
	ent.modTime = sys.time
	return &openFile{
		mu:   &sys.mu,
		ent:  ent,
		data: append([]byte(nil), ent.content...),
	}, nil
}

func (sys *System) Mkdir(ctx context.Context, path string, mode os.FileMode) error {
	wrap := pathErrorFunc("mkdir", path)
	path, err := cleanPath(path)
	if err != nil {
		return wrap(err)
	}

	defer sys.mu.Unlock()
	defer sys.stepTime()
	sys.mu.Lock()
	sys.init()
	_, err = sys.mkentry(path, os.ModeDir|mode&os.ModePerm)
	if err != nil {
		return wrap(err)
	}
	return nil
}

func (sys *System) Symlink(ctx context.Context, oldname, newname string) error {
	wrap := linkErrorFunc("symlink", oldname, newname)
	newname, err := cleanPath(newname)
	if err != nil {
		return wrap(err)
	}

	defer sys.mu.Unlock()
	defer sys.stepTime()
	sys.mu.Lock()
	sys.init()
	ent, err := sys.mkentry(newname, os.ModeSymlink|0777)
	if err != nil {
		return wrap(err)
	}
	ent.link = oldname
	return nil
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
	wrap := pathErrorFunc("remove", path)
	path, err := cleanPath(path)
	if err != nil {
		return wrap(err)
	}

	defer sys.mu.Unlock()
	defer sys.stepTime()
	sys.mu.Lock()
	sys.init()
	dir, name := filepath.Split(path)
	dir = sys.resolve(dir)
	par := sys.fs[dir]
	if par == nil || !par.mode.IsDir() {
		return wrap(os.ErrNotExist)
	}
	if par.mode&0222 == 0 {
		return wrap(os.ErrPermission)
	}
	path = filepath.Join(dir, name)
	ent := sys.fs[path]
	if ent == nil {
		return wrap(os.ErrNotExist)
	}
	if ent.mode.IsDir() && len(sys.readdir(path)) > 0 {
		return wrap(errors.New("fake OS: directory not empty"))
	}
	delete(sys.fs, path)
	return nil
}

// Mkprogram creates a filesystem entry that calls a program when run.
func (sys *System) Mkprogram(path string, prog Program) error {
	wrap := pathErrorFunc("mkprogram", path)
	path, err := cleanPath(path)
	if err != nil {
		return wrap(err)
	}

	defer sys.mu.Unlock()
	defer sys.stepTime()
	sys.mu.Lock()
	sys.init()
	ent, err := sys.mkentry(path, 0777)
	if err != nil {
		return wrap(err)
	}
	ent.program = prog
	return nil
}

func (sys *System) Run(ctx context.Context, cmd *system.Cmd) (output []byte, err error) {
	wrap := pathErrorFunc("exec", cmd.Path)
	path, err := cleanPath(cmd.Path)
	if err != nil {
		return nil, wrap(err)
	}

	sys.mu.Lock()
	sys.init()
	var (
		exists  bool
		mode    os.FileMode
		program Program
	)
	if ent := sys.fs[sys.resolve(path)]; ent != nil {
		exists = true
		mode = ent.mode
		program = ent.program
	}
	sys.stepTime()
	sys.mu.Unlock()

	if !exists {
		return nil, wrap(os.ErrNotExist)
	}
	if mode&0111 == 0 {
		return nil, wrap(os.ErrPermission)
	}
	if program == nil {
		return nil, wrap(errors.New("fake system: not a program"))
	}
	out := new(bytes.Buffer)
	exit := program(ctx, &ProgramContext{
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

func cleanPath(path string) (string, error) {
	if !filepath.IsAbs(path) {
		return "", errors.New("fake OS: path is not absolute")
	}
	return filepath.Clean(path), nil
}

func pathParts(path string) []string {
	if !filepath.IsAbs(path) {
		return nil
	}
	vol := filepath.VolumeName(path)
	path = vol + filepath.Clean(path[len(vol):])
	n := 1
	for p := path; ; {
		d := filepath.Dir(p)
		if p == d {
			break
		}
		n++
		p = d
	}
	parts := make([]string, n)
	for i, p := 0, path; i < n; i++ {
		parts[n-i-1] = filepath.Base(p)
		p = filepath.Dir(p)
	}
	return parts
}

func pathErrorFunc(op string, path string) func(error) error {
	return func(e error) error {
		if e == nil {
			return nil
		}
		return &os.PathError{Op: op, Path: path, Err: e}
	}
}

func linkErrorFunc(op string, oldname, newname string) func(error) error {
	return func(e error) error {
		if e == nil {
			return nil
		}
		return &os.LinkError{
			Op:  op,
			Old: oldname,
			New: newname,
			Err: e,
		}
	}
}

type openFile struct {
	data   []byte
	pos    int
	closed bool

	mu  *sync.Mutex // pointer to System.mu
	ent *entry
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
