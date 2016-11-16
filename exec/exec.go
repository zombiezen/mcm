package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/zombiezen/mcm/catalog"
	"github.com/zombiezen/mcm/internal/depgraph"
	"github.com/zombiezen/mcm/third_party/golang/capnproto"
)

func main() {
	c, err := readCatalog(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "mcm-exec: read catalog:", err)
		os.Exit(1)
	}
	res, _ := c.Resources()
	g, err := depgraph.New(res)
	if err != nil {
		fmt.Fprintln(os.Stderr, "mcm-exec:", err)
		os.Exit(1)
	}
	if err = applyCatalog(g); err != nil {
		fmt.Fprintln(os.Stderr, "mcm-exec:", err)
		os.Exit(1)
	}
}

func applyCatalog(g *depgraph.Graph) error {
	ok := true
	for !g.Done() {
		ready := g.Ready()
		if len(ready) == 0 {
			return errors.New("graph not done, but has nothing to do")
		}
		curr := ready[0]
		if err := applyResource(g.Resource(curr)); err == nil {
			g.Mark(curr)
		} else {
			// TODO(soon): log skipped resources
			g.MarkFailure(curr)
			fmt.Fprintf(os.Stderr, "mcm-exec: applying resource %d:", curr, err)
			ok = false
		}
	}
	if !ok {
		return errors.New("not all resources applied cleanly")
	}
	return nil
}

func applyResource(r catalog.Resource) error {
	wrap := func(e error) error {
		if e == nil {
			return nil
		}
		c, _ := r.Comment()
		if c == "" {
			return fmt.Errorf("apply %d: %v", r.ID(), e)
		}
		return fmt.Errorf("apply %q (id=%d): %v", c, r.ID(), e)
	}
	switch r.Which() {
	case catalog.Resource_Which_file:
		f, err := r.File()
		if err != nil {
			return wrap(err)
		}
		return wrap(applyFile(f))
	case catalog.Resource_Which_exec:
		e, err := r.Exec()
		if err != nil {
			return wrap(err)
		}
		return wrap(applyExec(e))
	default:
		return wrap(fmt.Errorf("unknown type %v", r.Which()))
	}
}

func applyFile(f catalog.File) error {
	path, err := f.Path()
	if err != nil {
		return fmt.Errorf("reading file path: %v", err)
	} else if path == "" {
		return errors.New("file path is empty")
	}
	fmt.Printf("considering file %s\n", path)
	switch f.Which() {
	case catalog.File_Which_plain:
		if f.Plain().HasContent() {
			content, _ := f.Plain().Content()
			// TODO(soon): respect file mode
			// TODO(soon): collect errors instead of stopping
			fmt.Printf("writing file %s\n", path)
			if err := ioutil.WriteFile(path, content, 0666); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unsupported file directive %v", f.Which())
	}
	return nil
}

func applyExec(e catalog.Exec) error {
	switch e.Condition().Which() {
	case catalog.Exec_condition_Which_always:
		// Continue.
	case catalog.Exec_condition_Which_onlyIf:
		cond, err := e.Condition().OnlyIf()
		if err != nil {
			return fmt.Errorf("condition: %v", err)
		}
		cmd, err := buildCommand(cond)
		if err != nil {
			return fmt.Errorf("condition: %v", err)
		}
		out, err := cmd.CombinedOutput()
		if _, exitFail := err.(*exec.ExitError); exitFail {
			return nil
		} else if err != nil {
			return fmt.Errorf("condition: %v; output:\n%s", err, out)
		}
	case catalog.Exec_condition_Which_unless:
		cond, err := e.Condition().Unless()
		if err != nil {
			return fmt.Errorf("condition: %v", err)
		}
		cmd, err := buildCommand(cond)
		if err != nil {
			return fmt.Errorf("condition: %v", err)
		}
		out, err := cmd.CombinedOutput()
		if err == nil {
			return nil
		} else if _, exitFail := err.(*exec.ExitError); !exitFail {
			return fmt.Errorf("condition: %v; output:\n%s", err, out)
		}
	case catalog.Exec_condition_Which_fileAbsent:
		path, _ := e.Condition().FileAbsent()
		if _, err := os.Lstat(path); err == nil {
			return nil
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("condition: %v", err)
		}
	default:
		return fmt.Errorf("unknown condition %v", e.Condition().Which())
	}

	main, err := e.Command()
	if err != nil {
		return fmt.Errorf("command: %v", err)
	}
	cmd, err := buildCommand(main)
	if err != nil {
		return fmt.Errorf("command: %v", err)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command: %v; output:\n%s", err, out)
	}
	return nil
}

func buildCommand(cmd catalog.Exec_Command) (*exec.Cmd, error) {
	var c *exec.Cmd
	switch cmd.Which() {
	case catalog.Exec_Command_Which_argv:
		argList, _ := cmd.Argv()
		if argList.Len() == 0 {
			return nil, fmt.Errorf("0-length argv")
		}
		argv := make([]string, argList.Len())
		for i := range argv {
			var err error
			argv[i], err = argList.At(i)
			if err != nil {
				return nil, fmt.Errorf("argv[%d]: %v", i, err)
			}
		}
		if !filepath.IsAbs(argv[0]) {
			return nil, fmt.Errorf("argv[0] (%q) is not an absolute path", argv[0])
		}
		c = &exec.Cmd{
			Path: argv[0],
			Args: argv,
		}
	default:
		return nil, fmt.Errorf("unsupported command type %v", cmd.Which())
	}

	env, _ := cmd.Environment()
	c.Env = make([]string, env.Len())
	for i := range c.Env {
		ei := env.At(i)
		k, err := ei.NameBytes()
		if err != nil {
			return nil, fmt.Errorf("getting environment[%d]: %v", i, err)
		} else if len(k) == 0 {
			return nil, fmt.Errorf("environment[%d] missing name", i)
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
		return nil, fmt.Errorf("working directory %q is not absolute", c.Dir)
	}

	return c, nil
}

func readCatalog(r io.Reader) (catalog.Catalog, error) {
	msg, err := capnp.NewDecoder(r).Decode()
	if err != nil {
		return catalog.Catalog{}, err
	}
	return catalog.ReadRootCatalog(msg)
}
