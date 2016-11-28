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

package system

import (
	"context"
	"os"
	"os/exec"
)

// Local implements FS and Runner by calling to the os package.
type Local struct{}

var _ System = Local{}

// Lstat calls os.Lstat.
func (Local) Lstat(ctx context.Context, path string) (os.FileInfo, error) {
	return os.Lstat(path)
}

// Mkdir calls os.Mkdir.
func (Local) Mkdir(ctx context.Context, path string, mode os.FileMode) error {
	return os.Mkdir(path, mode)
}

// Remove calls os.Remove.
func (Local) Remove(ctx context.Context, path string) error {
	return os.Remove(path)
}

// Symlink calls os.Symlink.
func (Local) Symlink(ctx context.Context, oldname, newname string) error {
	return os.Symlink(oldname, newname)
}

// CreateFile calls os.OpenFile with write-only and exclusive create flags.
func (Local) CreateFile(ctx context.Context, path string, mode os.FileMode) (FileWriter, error) {
	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
}

// OpenFile calls os.OpenFile with read-write.
func (Local) OpenFile(ctx context.Context, path string) (File, error) {
	return os.OpenFile(path, os.O_RDWR, 0666)
}

// Run runs a process using os/exec and returns the combined stdout and stderr.
func (Local) Run(ctx context.Context, cmd *Cmd) (output []byte, err error) {
	ec := &exec.Cmd{
		Path: cmd.Path,
		Args: cmd.Args,
		Env:  cmd.Env,
		Dir:  cmd.Dir,
	}
	return ec.CombinedOutput()
}
