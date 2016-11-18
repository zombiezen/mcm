package execlib

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/zombiezen/mcm/catalog"
	"github.com/zombiezen/mcm/internal/depgraph"
)

type Applier struct {
	Input io.Reader
	Log   Logger
	OS    OS

	// Unconditional skips evaluating any conditions, assuming that
	// everything must be applied.
	Unconditional bool
}

type Logger interface {
	Infof(ctx context.Context, format string, args ...interface{})
	Error(ctx context.Context, err error)
}

type OS interface {
	Lstat(path string) (os.FileInfo, error)
	WriteFile(path string, content []byte, mode os.FileMode) error
	Mkdir(path string, mode os.FileMode) error
	Remove(path string) error
	Run(context.Context, *exec.Cmd) (output []byte, err error)
}

func (app *Applier) Apply(ctx context.Context, c catalog.Catalog) error {
	res, _ := c.Resources()
	g, err := depgraph.New(res)
	if err != nil {
		return toError(err)
	}
	if err = app.applyCatalog(ctx, g); err != nil {
		return toError(err)
	}
	return nil
}

func (app *Applier) applyCatalog(ctx context.Context, g *depgraph.Graph) error {
	ok := true
	for !g.Done() {
		ready := g.Ready()
		if len(ready) == 0 {
			return errors.New("graph not done, but has nothing to do")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		curr := ready[0]
		res := g.Resource(curr)
		app.Log.Infof(ctx, "applying: %s", formatResource(res))
		if err := errorWithResource(res, app.applyResource(ctx, res)); err == nil {
			g.Mark(curr)
		} else {
			ok = false
			app.Log.Error(ctx, toError(err).(*Error))
			skipped := g.MarkFailure(curr)
			if len(skipped) > 0 {
				skipnames := make([]string, len(skipped))
				for i := range skipnames {
					skipnames[i] = formatResource(g.Resource(skipped[i]))
				}
				app.Log.Infof(ctx, "skipping due to failure of %s: %s", formatResource(res), strings.Join(skipnames, ", "))
			}
		}
	}
	if !ok {
		return errors.New("not all resources applied cleanly")
	}
	return nil
}

func (app *Applier) applyResource(ctx context.Context, r catalog.Resource) error {
	switch r.Which() {
	case catalog.Resource_Which_noop:
		return nil
	case catalog.Resource_Which_file:
		f, err := r.File()
		if err != nil {
			return err
		}
		return app.applyFile(ctx, f)
	case catalog.Resource_Which_exec:
		e, err := r.Exec()
		if err != nil {
			return err
		}
		return app.applyExec(ctx, e)
	default:
		return errorf("unknown type %v", r.Which())
	}
}

func (app *Applier) applyFile(ctx context.Context, f catalog.File) error {
	path, err := f.Path()
	if err != nil {
		return errorf("read file path from catalog: %v", err)
	} else if path == "" {
		return errors.New("file path is empty")
	}
	switch f.Which() {
	case catalog.File_Which_plain:
		if f.Plain().HasContent() {
			content, err := f.Plain().Content()
			if err != nil {
				return errorf("read content from catalog: %v", err)
			}
			// TODO(soon): respect file mode
			if err := app.OS.WriteFile(path, content, 0666); err != nil {
				return err
			}
		} else {
			info, err := app.OS.Lstat(path)
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				// TODO(soon): what kind of node it?
				return errorf("%s is not a regular file")
			}
		}
	case catalog.File_Which_directory:
		// TODO(soon): respect file mode
		if err := app.OS.Mkdir(path, 0777); err == nil || !os.IsExist(err) {
			return err
		}
		// Ensure that what exists is a directory.
		info, err := app.OS.Lstat(path)
		if err != nil {
			return errorf("determine state of %s: %v", path, err)
		}
		if !info.IsDir() {
			// TODO(soon): what kind of node it?
			return errorf("%s is not a directory", path)
		}
	case catalog.File_Which_absent:
		err := app.OS.Remove(path)
		if err == nil || !os.IsNotExist(err) {
			return err
		}
	default:
		return errorf("unsupported file directive %v", f.Which())
	}
	return nil
}

func (app *Applier) applyExec(ctx context.Context, e catalog.Exec) error {
	if !app.Unconditional {
		switch e.Condition().Which() {
		case catalog.Exec_condition_Which_always:
			// Continue.
		case catalog.Exec_condition_Which_onlyIf:
			cond, err := e.Condition().OnlyIf()
			if err != nil {
				return errorf("condition: %v", err)
			}
			cmd, err := buildCommand(cond)
			if err != nil {
				return errorf("condition: %v", err)
			}
			out, err := app.OS.Run(ctx, cmd)
			if _, exitFail := err.(*exec.ExitError); exitFail {
				return nil
			} else if err != nil {
				return errorWithOutput(out, errorf("condition: %v", err))
			}
		case catalog.Exec_condition_Which_unless:
			cond, err := e.Condition().Unless()
			if err != nil {
				return errorf("condition: %v", err)
			}
			cmd, err := buildCommand(cond)
			if err != nil {
				return errorf("condition: %v", err)
			}
			out, err := app.OS.Run(ctx, cmd)
			if err == nil {
				return nil
			} else if _, exitFail := err.(*exec.ExitError); !exitFail {
				return errorWithOutput(out, errorf("condition: %v", err))
			}
		case catalog.Exec_condition_Which_fileAbsent:
			path, _ := e.Condition().FileAbsent()
			if _, err := app.OS.Lstat(path); err == nil {
				// File exists; skip command.
				return nil
			} else if !os.IsNotExist(err) {
				return errorf("condition: %v", err)
			}
		default:
			return errorf("unknown condition %v", e.Condition().Which())
		}
	}

	main, err := e.Command()
	if err != nil {
		return errorf("command: %v", err)
	}
	cmd, err := buildCommand(main)
	if err != nil {
		return errorf("command: %v", err)
	}
	out, err := app.OS.Run(ctx, cmd)
	if err != nil {
		return errorWithOutput(out, errorf("command: %v", err))
	}
	return nil
}

func buildCommand(cmd catalog.Exec_Command) (*exec.Cmd, error) {
	var c *exec.Cmd
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
		c = &exec.Cmd{
			Path: argv[0],
			Args: argv,
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
		// TODO(windows): conditionally use "C:\"
		c.Dir = "/"
	} else if !filepath.IsAbs(c.Dir) {
		return nil, errorf("working directory %q is not absolute", c.Dir)
	}

	return c, nil
}

func formatResource(r catalog.Resource) string {
	c, _ := r.Comment()
	if c == "" {
		return fmt.Sprintf("id=%d", r.ID())
	}
	return fmt.Sprintf("%s (id=%d)", c, r.ID())
}
