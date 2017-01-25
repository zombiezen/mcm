// Copyright 2017 The Minimal Configuration Manager Authors
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
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/zombiezen/mcm/catalog"
	"github.com/zombiezen/mcm/internal/system"
)

type job struct {
	sys         system.System
	log         Logger
	resource    catalog.Resource
	depsChanged map[uint64]bool

	bashPath string
}

type jobResult struct {
	id      uint64
	changed bool
	err     error
}

func (j *job) run(ctx context.Context) jobResult {
	result := jobResult{id: j.resource.ID()}
	switch j.resource.Which() {
	case catalog.Resource_Which_noop:
		for _, c := range j.depsChanged {
			if c {
				result.changed = true
				break
			}
		}
		return result
	case catalog.Resource_Which_file:
		f, err := j.resource.File()
		if err != nil {
			result.err = errorWithResource(j.resource, err)
			return result
		}
		changed, err := j.file(ctx, f)
		if err != nil {
			result.err = errorWithResource(j.resource, err)
			return result
		}
		result.changed = changed
		return result
	case catalog.Resource_Which_exec:
		e, err := j.resource.Exec()
		if err != nil {
			result.err = errorWithResource(j.resource, err)
			return result
		}
		changed, err := j.exec(ctx, e)
		if err != nil {
			result.err = errorWithResource(j.resource, err)
			return result
		}
		result.changed = changed
		return result
	default:
		result.err = errorWithResource(j.resource, errorf("unknown type %v", j.resource.Which()))
		return result
	}
}

func (j *job) file(ctx context.Context, f catalog.File) (changed bool, err error) {
	path, err := f.Path()
	if err != nil {
		return false, errorf("read file path from catalog: %v", err)
	}
	if path == "" {
		return false, errors.New("file path is empty")
	}
	switch f.Which() {
	case catalog.File_Which_plain:
		return j.plainFile(ctx, path, f.Plain())
	case catalog.File_Which_directory:
		return j.directory(ctx, path, f.Directory())
	case catalog.File_Which_symlink:
		return j.symlink(ctx, path, f.Symlink())
	case catalog.File_Which_absent:
		err := j.sys.Remove(ctx, path)
		if err != nil {
			if os.IsNotExist(err) {
				return false, nil
			}
			return false, err
		}
		return true, nil
	default:
		return false, errorf("unsupported file directive %v", f.Which())
	}
}

func (j *job) plainFile(ctx context.Context, path string, f catalog.File_plain) (changed bool, err error) {
	if !f.HasContent() {
		info, err := j.sys.Lstat(ctx, path)
		if err != nil {
			return false, err
		}
		if !info.Mode().IsRegular() {
			// TODO(soon): what kind of node it?
			return false, errorf("%s is not a regular file")
		}
		mode, _ := f.Mode()
		return j.fileModeWithInfo(ctx, path, info, mode)
	}

	content, err := f.Content()
	if err != nil {
		return false, errorf("read content from catalog: %v", err)
	}
	contentChanged, err := j.plainFileContent(ctx, path, content)
	if err != nil {
		return false, err
	}
	mode, _ := f.Mode()
	modeChanged, err := j.fileMode(ctx, path, mode)
	if err != nil {
		return false, err
	}
	return contentChanged || modeChanged, nil
}

func (j *job) plainFileContent(ctx context.Context, path string, content []byte) (changed bool, err error) {
	w, err := j.sys.CreateFile(ctx, path, 0666) // rely on umask to restrict
	if os.IsExist(err) {
		f, err := j.sys.OpenFile(ctx, path)
		if err != nil {
			return false, err
		}
		matches, err := hasContent(f, content)
		if err != nil {
			f.Close()
			return false, err
		}
		if matches {
			f.Close()
			return false, nil
		}
		if _, err = f.Seek(0, io.SeekStart); err != nil {
			f.Close()
			return false, err
		}
		if err = f.Truncate(0); err != nil {
			f.Close()
			return false, err
		}
		w = f
	} else if err != nil {
		return false, err
	}
	_, err = w.Write(content)
	cerr := w.Close()
	if err != nil {
		return false, err
	}
	if cerr != nil {
		return false, cerr
	}
	return true, nil
}

func hasContent(r io.Reader, content []byte) (bool, error) {
	r = &errReader{r: r}
	buf := make([]byte, 4096)
	for len(content) > 0 {
		n, err := r.Read(buf)
		if n > len(content) || !bytes.Equal(buf[:n], content[:n]) {
			return false, nil
		}
		content = content[n:]
		if err == io.EOF {
			return len(content) == 0, nil
		}
		if err != nil {
			return false, err
		}
	}
	n, err := r.Read(buf)
	if n > 0 {
		return false, nil
	}
	if err != io.EOF {
		return false, err
	}
	return true, nil
}

type errReader struct {
	r   io.Reader
	err error
}

func (er *errReader) Read(p []byte) (n int, _ error) {
	if er.err != nil {
		return 0, er.err
	}
	n, er.err = er.r.Read(p)
	return n, er.err
}

func (j *job) directory(ctx context.Context, path string, d catalog.File_directory) (changed bool, err error) {
	err = j.sys.Mkdir(ctx, path, 0777) // rely on umask to restrict
	if err == nil {
		mode, _ := d.Mode()
		_, err = j.fileMode(ctx, path, mode)
		return true, err
	}
	if !os.IsExist(err) {
		return false, err
	}
	// Ensure that what exists is a directory.
	info, err := j.sys.Lstat(ctx, path)
	if err != nil {
		return false, errorf("determine state of %s: %v", path, err)
	}
	if !info.IsDir() {
		// TODO(soon): what kind of node it?
		return false, errorf("%s is not a directory", path)
	}
	mode, _ := d.Mode()
	return j.fileModeWithInfo(ctx, path, info, mode)
}

func (j *job) symlink(ctx context.Context, path string, l catalog.File_symlink) (changed bool, err error) {
	target, err := l.Target()
	if err != nil {
		return false, errorf("read target from catalog: %v", err)
	}
	err = j.sys.Symlink(ctx, target, path)
	if err == nil {
		return true, nil
	}
	if !os.IsExist(err) {
		return false, err
	}
	// Ensure that what exists is a symlink before trying to retarget.
	info, err := j.sys.Lstat(ctx, path)
	if err != nil {
		return false, errorf("determine state of %s: %v", path, err)
	}
	if info.Mode()&os.ModeType != os.ModeSymlink {
		// TODO(soon): what kind of node is it?
		return false, errorf("%s is not a symlink", path)
	}
	actual, err := j.sys.Readlink(ctx, path)
	if err != nil {
		return false, err
	}
	if actual == target {
		// Already the correct link.
		return false, nil
	}
	if err := j.sys.Remove(ctx, path); err != nil {
		return false, errorf("retargeting %s: %v", path, err)
	}
	if err := j.sys.Symlink(ctx, target, path); err != nil {
		return false, errorf("retargeting %s: %v", path, err)
	}
	return true, nil
}

func (j *job) fileMode(ctx context.Context, path string, mode catalog.File_Mode) (changed bool, err error) {
	// TODO(someday): avoid the extra capnp read, since WithInfo also accesses these fields.
	bits := mode.Bits()
	user, _ := mode.User()
	group, _ := mode.Group()
	if bits == catalog.File_Mode_unset && isZeroUserRef(user) && isZeroGroupRef(group) {
		return false, nil
	}
	st, err := j.sys.Lstat(ctx, path)
	if err != nil {
		return false, err
	}
	return j.fileModeWithInfo(ctx, path, st, mode)
}

func (j *job) fileModeWithInfo(ctx context.Context, path string, st os.FileInfo, mode catalog.File_Mode) (changed bool, err error) {
	bits := mode.Bits()
	user, _ := mode.User()
	group, _ := mode.Group()
	changedBits, err := j.fileModeBits(ctx, path, st, bits)
	if err != nil {
		return false, err
	}
	changedOwner, err := j.fileModeOwner(ctx, path, st, user, group)
	if err != nil {
		return false, err
	}
	return changedBits || changedOwner, nil
}

func (j *job) fileModeBits(ctx context.Context, path string, info os.FileInfo, bits uint16) (changed bool, err error) {
	if bits == catalog.File_Mode_unset {
		return false, nil
	}
	newMode := modeFromCatalog(bits)
	const mask = os.ModePerm | os.ModeSticky | os.ModeSetuid | os.ModeSetgid
	if info.Mode()&mask == newMode {
		return false, nil
	}
	if err := j.sys.Chmod(ctx, path, newMode); err != nil {
		return false, err
	}
	return true, nil
}

func (j *job) fileModeOwner(ctx context.Context, path string, info os.FileInfo, user catalog.UserRef, group catalog.GroupRef) (changed bool, err error) {
	uid, err := resolveUserRef(j.sys, user)
	if err != nil {
		return false, errorf("resolve user: %v", err)
	}
	gid, err := resolveGroupRef(j.sys, group)
	if err != nil {
		return false, errorf("resolve group: %v", err)
	}
	if uid == -1 && gid == -1 {
		return false, nil
	}
	if oldUID, oldGID, err := j.sys.OwnerInfo(info); err != nil {
		j.log.Infof(ctx, "%s: reading file owner: %v; assuming need to chown", formatResource(j.resource), err)
	} else if (uid == -1 || oldUID == uid) && (gid == -1 || oldGID == gid) {
		return false, nil
	}
	if err := j.sys.Chown(ctx, path, uid, gid); err != nil {
		return false, err
	}
	return true, nil
}

func resolveUserRef(lookup system.UserLookup, ref catalog.UserRef) (system.UID, error) {
	switch ref.Which() {
	case catalog.UserRef_Which_ID:
		id := ref.ID()
		if id < -1 {
			return -1, fmt.Errorf("invalid uid %d", id)
		}
		return system.UID(id), nil
	case catalog.UserRef_Which_name:
		name, err := ref.Name()
		if err != nil {
			return -1, err
		}
		return lookup.LookupUser(name)
	default:
		return -1, fmt.Errorf("unhandled user ref type %v", ref.Which())
	}
}

func resolveGroupRef(lookup system.UserLookup, ref catalog.GroupRef) (system.GID, error) {
	switch ref.Which() {
	case catalog.GroupRef_Which_ID:
		id := ref.ID()
		if id < -1 {
			return -1, fmt.Errorf("invalid gid %d", id)
		}
		return system.GID(id), nil
	case catalog.GroupRef_Which_name:
		name, err := ref.Name()
		if err != nil {
			return -1, err
		}
		return lookup.LookupGroup(name)
	default:
		return -1, fmt.Errorf("unhandled group ref type %v", ref.Which())
	}
}

func isZeroUserRef(ref catalog.UserRef) bool {
	return ref.Which() == catalog.UserRef_Which_ID && ref.ID() == -1
}

func isZeroGroupRef(ref catalog.GroupRef) bool {
	return ref.Which() == catalog.GroupRef_Which_ID && ref.ID() == -1
}

func modeFromCatalog(cmode uint16) os.FileMode {
	m := os.FileMode(cmode & catalog.File_Mode_permMask)
	if cmode&catalog.File_Mode_sticky != 0 {
		m |= os.ModeSticky
	}
	if cmode&catalog.File_Mode_setuid != 0 {
		m |= os.ModeSetuid
	}
	if cmode&catalog.File_Mode_setgid != 0 {
		m |= os.ModeSetgid
	}
	return m
}

func (j *job) exec(ctx context.Context, e catalog.Exec) (changed bool, err error) {
	proceed, err := j.evalExecCondition(ctx, e.Condition())
	if err != nil {
		return false, errorf("condition: %v", err)
	}
	if !proceed {
		return false, nil
	}
	cmd, err := e.Command()
	if err != nil {
		return false, errorf("command: %v", err)
	}
	if err := j.runCommand(ctx, cmd); err != nil {
		return false, errorf("command: %v", err)
	}
	return true, nil
}

func (j *job) evalExecCondition(ctx context.Context, cond catalog.Exec_condition) (proceed bool, err error) {
	switch cond.Which() {
	case catalog.Exec_condition_Which_always:
		return true, nil
	case catalog.Exec_condition_Which_onlyIf:
		c, err := cond.OnlyIf()
		if err != nil {
			return false, err
		}
		return j.runCondition(ctx, c)
	case catalog.Exec_condition_Which_unless:
		c, err := cond.Unless()
		if err != nil {
			return false, err
		}
		success, err := j.runCondition(ctx, c)
		if err != nil {
			return false, err
		}
		return !success, nil
	case catalog.Exec_condition_Which_fileAbsent:
		path, _ := cond.FileAbsent()
		_, err := j.sys.Lstat(ctx, path)
		if err != nil {
			if os.IsNotExist(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	case catalog.Exec_condition_Which_ifDepsChanged:
		deps, err := cond.IfDepsChanged()
		if err != nil {
			return false, err
		}
		n := deps.Len()
		if n == 0 {
			return false, errorf("ifDepsChanged is empty list")
		}
		for i := 0; i < n; i++ {
			id := deps.At(i)
			if _, ok := j.depsChanged[id]; !ok {
				return false, errorf("depends on ID %d, which is not in resource's direct dependencies", id)
			}
		}
		for i := 0; i < n; i++ {
			if j.depsChanged[deps.At(i)] {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, errorf("unknown condition %v", cond.Which())
	}
}

func (j *job) runCommand(ctx context.Context, c catalog.Exec_Command) error {
	cmd, err := buildCommand(c, j.bashPath)
	if err != nil {
		return err
	}
	out, err := j.sys.Run(ctx, cmd)
	if err != nil {
		return errorWithOutput(out, err)
	}
	return nil
}

func (j *job) runCondition(ctx context.Context, c catalog.Exec_Command) (success bool, err error) {
	cmd, err := buildCommand(c, j.bashPath)
	if err != nil {
		return false, err
	}
	out, err := j.sys.Run(ctx, cmd)
	if _, fail := err.(*exec.ExitError); fail {
		return false, nil
	}
	if err != nil {
		return false, errorWithOutput(out, err)
	}
	return true, nil
}

func buildCommand(cmd catalog.Exec_Command, bashPath string) (*system.Cmd, error) {
	var c *system.Cmd
	switch cmd.Which() {
	case catalog.Exec_Command_Which_argv:
		argList, _ := cmd.Argv()
		if argList.Len() == 0 {
			return nil, errorf("0-length argv")
		}
		argv := make([]string, argList.Len())
		for i := range argv {
			var err error
			argv[i], err = argList.At(i)
			if err != nil {
				return nil, errorf("argv[%d]: %v", i, err)
			}
		}
		if !filepath.IsAbs(argv[0]) {
			return nil, errorf("argv[0] (%q) is not an absolute path", argv[0])
		}
		c = &system.Cmd{
			Path: argv[0],
			Args: argv,
		}
	case catalog.Exec_Command_Which_bash:
		b, err := cmd.BashBytes()
		if err != nil {
			return nil, errorf("read bash: %v", err)
		}
		c = &system.Cmd{
			Path:  bashPath,
			Args:  []string{bashPath},
			Stdin: bytes.NewReader(b),
		}
	default:
		return nil, errorf("unsupported command type %v", cmd.Which())
	}

	env, _ := cmd.Environment()
	c.Env = make([]string, env.Len())
	for i := range c.Env {
		ei := env.At(i)
		k, err := ei.NameBytes()
		if err != nil {
			return nil, errorf("getting environment[%d]: %v", i, err)
		} else if len(k) == 0 {
			return nil, errorf("environment[%d] missing name", i)
		}
		v, _ := ei.ValueBytes()
		buf := make([]byte, 0, len(k)+len(v)+1)
		buf = append(buf, k...)
		buf = append(buf, '=')
		buf = append(buf, v...)
		c.Env[i] = string(buf)
	}

	c.Dir, _ = cmd.WorkingDirectory()
	if c.Dir == "" {
		c.Dir = system.LocalRoot
	} else if !filepath.IsAbs(c.Dir) {
		return nil, errorf("working directory %q is not absolute", c.Dir)
	}

	return c, nil
}
