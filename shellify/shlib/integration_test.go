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

package shlib_test

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zombiezen/mcm/catalog"
	"github.com/zombiezen/mcm/internal/catpogs"
	"github.com/zombiezen/mcm/shellify/shlib"
)

var keepScripts = flag.Bool("keep_scripts", false, "do not remove generated scripts from temporary directory")

func TestIntegration(t *testing.T) {
	bashPath, err := exec.LookPath("bash")
	if err != nil {
		t.Skipf("Can't find bash: %v", err)
	}
	t.Logf("using %s for bash", bashPath)
	t.Run("Empty", func(t *testing.T) { emptyTest(t, bashPath) })
	t.Run("File", func(t *testing.T) { fileTest(t, bashPath) })
	t.Run("Noop", func(t *testing.T) { noopTest(t, bashPath) })
	t.Run("NoContentFile", func(t *testing.T) { noContentFileTest(t, bashPath) })
	t.Run("Link", func(t *testing.T) { linkTest(t, bashPath) })
	t.Run("Relink", func(t *testing.T) { relinkTest(t, bashPath) })
	t.Run("SkipFail", func(t *testing.T) { skipFailTest(t, bashPath) })
	t.Run("Exec", func(t *testing.T) { execTest(t, bashPath) })
	t.Run("ExecBash", func(t *testing.T) { execBashTest(t, bashPath) })
	t.Run("ExecIfDepsChanged", func(t *testing.T) { execIfDepsChangedTest(t, bashPath) })
}

func emptyTest(t *testing.T, bashPath string) {
	c, err := new(catpogs.Catalog).ToCapnp()
	if err != nil {
		t.Fatalf("build empty catalog: %v", err)
	}
	_, err = runCatalog("empty", bashPath, t, c)
	if err != nil {
		t.Errorf("run catalog: %v", err)
	}
}

func fileTest(t *testing.T, bashPath string) {
	root, deleteTempDir, err := makeTempDir(t)
	if err != nil {
		t.Fatalf("temp directory: %v", err)
	}
	defer deleteTempDir()
	fpath := filepath.Join(root, "foo.txt")
	const fileContent = "Hello!\n"
	c, err := (&catpogs.Catalog{
		Resources: []*catpogs.Resource{
			{
				ID:      42,
				Comment: "file",
				Which:   catalog.Resource_Which_file,
				File:    catpogs.PlainFile(fpath, []byte(fileContent)),
			},
		},
	}).ToCapnp()
	if err != nil {
		t.Fatalf("build catalog: %v", err)
	}
	_, err = runCatalog("file", bashPath, t, c)
	if err != nil {
		t.Errorf("run catalog: %v", err)
	}
	gotContent, err := ioutil.ReadFile(fpath)
	if err != nil {
		t.Errorf("read %s: %v", fpath, err)
	}
	if !bytes.Equal(gotContent, []byte(fileContent)) {
		t.Errorf("content of %s = %q; want %q", fpath, gotContent, fileContent)
	}
}

func noopTest(t *testing.T, bashPath string) {
	t.Run("NoDeps", func(t *testing.T) {
		c, err := (&catpogs.Catalog{
			Resources: []*catpogs.Resource{
				{
					ID:      42,
					Comment: "noop",
					Which:   catalog.Resource_Which_noop,
				},
			},
		}).ToCapnp()
		if err != nil {
			t.Fatalf("build catalog: %v", err)
		}
		_, err = runCatalog("noopNoDeps", bashPath, t, c)
		if err != nil {
			t.Errorf("run catalog: %v", err)
		}
	})
	t.Run("WithDeps", func(t *testing.T) {
		root, deleteTempDir, err := makeTempDir(t)
		if err != nil {
			t.Fatalf("temp directory: %v", err)
		}
		defer deleteTempDir()
		fpath1 := filepath.Join(root, "foo1.txt")
		fpath2 := filepath.Join(root, "foo2.txt")
		c, err := (&catpogs.Catalog{
			Resources: []*catpogs.Resource{
				{
					ID:      101,
					Comment: "file1",
					Which:   catalog.Resource_Which_file,
					File:    catpogs.PlainFile(fpath1, []byte("Hello!\n")),
				},
				{
					ID:      102,
					Comment: "file2",
					Which:   catalog.Resource_Which_file,
					File:    catpogs.PlainFile(fpath2, []byte("Hello!\n")),
				},
				{
					ID:      42,
					Comment: "noop",
					Deps:    []uint64{101, 102},
					Which:   catalog.Resource_Which_noop,
				},
			},
		}).ToCapnp()
		if err != nil {
			t.Fatalf("build catalog: %v", err)
		}
		_, err = runCatalog("noopWithDeps", bashPath, t, c)
		if err != nil {
			t.Errorf("run catalog: %v", err)
		}
	})
	t.Run("WithDepsChange", func(t *testing.T) {
		root, deleteTempDir, err := makeTempDir(t)
		if err != nil {
			t.Fatalf("temp directory: %v", err)
		}
		defer deleteTempDir()
		fpath1 := filepath.Join(root, "foo1.txt")
		fpath2 := filepath.Join(root, "foo2.txt")
		outpath := filepath.Join(root, "canary.txt")
		c, err := (&catpogs.Catalog{
			Resources: []*catpogs.Resource{
				{
					ID:      101,
					Comment: "file1",
					Which:   catalog.Resource_Which_file,
					File:    catpogs.PlainFile(fpath1, []byte("Hello!\n")),
				},
				{
					ID:      102,
					Comment: "file2",
					Which:   catalog.Resource_Which_file,
					File:    catpogs.PlainFile(fpath2, []byte("Hello!\n")),
				},
				{
					ID:      42,
					Comment: "noop",
					Deps:    []uint64{101, 102},
					Which:   catalog.Resource_Which_noop,
				},
				{
					ID:      200,
					Comment: "canary",
					Deps:    []uint64{42},
					Which:   catalog.Resource_Which_exec,
					Exec: &catpogs.Exec{
						Command: &catpogs.Command{
							Which: catalog.Exec_Command_Which_argv,
							Argv:  []string{"/usr/bin/touch", outpath},
						},
						Condition: catpogs.ExecCondition{
							Which:         catalog.Exec_condition_Which_ifDepsChanged,
							IfDepsChanged: []uint64{42},
						},
					},
				},
			},
		}).ToCapnp()
		if err != nil {
			t.Fatalf("build catalog: %v", err)
		}
		_, err = runCatalog("noopWithDepsChange", bashPath, t, c)
		if err != nil {
			t.Errorf("run catalog: %v", err)
		}
		if _, err := os.Lstat(outpath); os.IsNotExist(err) {
			t.Errorf("%q does not exist; noop resource did not report changed", outpath)
		} else if err != nil {
			t.Errorf("os.Lstat(%q) error: %v", outpath, err)
		}
	})
}

func noContentFileTest(t *testing.T, bashPath string) {
	t.Run("Exists", func(t *testing.T) {
		root, deleteTempDir, err := makeTempDir(t)
		if err != nil {
			t.Fatalf("temp directory: %v", err)
		}
		defer deleteTempDir()
		fpath := filepath.Join(root, "foo.txt")
		if err := ioutil.WriteFile(fpath, []byte("data\n"), 0666); err != nil {
			t.Fatal("WriteFile:", err)
		}
		c, err := (&catpogs.Catalog{
			Resources: []*catpogs.Resource{
				{
					ID:      42,
					Comment: "file",
					Which:   catalog.Resource_Which_file,
					File:    catpogs.PlainFile(fpath, nil),
				},
			},
		}).ToCapnp()
		if err != nil {
			t.Fatalf("build catalog: %v", err)
		}
		_, err = runCatalog("noContentExists", bashPath, t, c)
		if err != nil {
			t.Errorf("run catalog: %v", err)
		}
	})
	t.Run("NotExists", func(t *testing.T) {
		root, deleteTempDir, err := makeTempDir(t)
		if err != nil {
			t.Fatalf("temp directory: %v", err)
		}
		defer deleteTempDir()
		fpath := filepath.Join(root, "foo.txt")
		c, err := (&catpogs.Catalog{
			Resources: []*catpogs.Resource{
				{
					ID:      42,
					Comment: "file",
					Which:   catalog.Resource_Which_file,
					File:    catpogs.PlainFile(fpath, nil),
				},
			},
		}).ToCapnp()
		if err != nil {
			t.Fatalf("build catalog: %v", err)
		}
		_, err = runCatalog("noContentNotExists", bashPath, t, c)
		if err == nil {
			t.Error("run catalog did not fail as expected")
		}
	})
}

func linkTest(t *testing.T, bashPath string) {
	root, deleteTempDir, err := makeTempDir(t)
	if err != nil {
		t.Fatalf("temp directory: %v", err)
	}
	defer deleteTempDir()
	fpath := filepath.Join(root, "foo")
	lpath := filepath.Join(root, "link")
	c, err := (&catpogs.Catalog{
		Resources: []*catpogs.Resource{
			{
				ID:      42,
				Comment: "file",
				Which:   catalog.Resource_Which_file,
				File:    catpogs.PlainFile(fpath, []byte("Hello")),
			},
			{
				ID:      100,
				Deps:    []uint64{42},
				Comment: "link",
				Which:   catalog.Resource_Which_file,
				File:    catpogs.SymlinkFile(fpath, lpath),
			},
		},
	}).ToCapnp()
	if err != nil {
		t.Fatalf("build catalog: %v", err)
	}
	_, err = runCatalog("link", bashPath, t, c)
	if err != nil {
		t.Errorf("run catalog: %v", err)
	}

	if info, err := os.Lstat(lpath); err == nil {
		if info.Mode()&os.ModeType != os.ModeSymlink {
			t.Errorf("os.Lstat(%q).Mode() = %v; want symlink", lpath, info.Mode())
		}
	} else {
		t.Errorf("os.Lstat(%q): %v", lpath, err)
	}
	if target, err := os.Readlink(lpath); err == nil {
		if target != fpath {
			t.Errorf("os.Readlink(%q) = %q; want %q", lpath, target, fpath)
		}
	} else {
		t.Errorf("os.Readlink(%q): %v", lpath, err)
	}
}

func relinkTest(t *testing.T, bashPath string) {
	root, deleteTempDir, err := makeTempDir(t)
	if err != nil {
		t.Fatalf("temp directory: %v", err)
	}
	defer deleteTempDir()
	f1path := filepath.Join(root, "foo")
	f2path := filepath.Join(root, "bar")
	lpath := filepath.Join(root, "link")
	c, err := (&catpogs.Catalog{
		Resources: []*catpogs.Resource{
			{
				ID:      42,
				Comment: "link",
				Which:   catalog.Resource_Which_file,
				File:    catpogs.SymlinkFile(f2path, lpath),
			},
		},
	}).ToCapnp()
	if err != nil {
		t.Fatalf("build catalog: %v", err)
	}
	if err := ioutil.WriteFile(f1path, []byte("File 1"), 0666); err != nil {
		t.Fatal("WriteFile 1:", err)
	}
	if err := ioutil.WriteFile(f2path, []byte("File 2"), 0666); err != nil {
		t.Fatal("WriteFile 2:", err)
	}
	if err := os.Symlink(f1path, lpath); err != nil {
		t.Fatalf("os.Symlink %s -> %s: %v", lpath, f1path, err)
	}
	_, err = runCatalog("relink", bashPath, t, c)
	if err != nil {
		t.Errorf("run catalog: %v", err)
	}

	if info, err := os.Lstat(lpath); err == nil {
		if info.Mode()&os.ModeType != os.ModeSymlink {
			t.Errorf("os.Lstat(%q).Mode() = %v; want symlink", lpath, info.Mode())
		}
	} else {
		t.Errorf("os.Lstat(%q): %v", lpath, err)
	}
	if target, err := os.Readlink(lpath); err == nil {
		if target != f2path {
			t.Errorf("os.Readlink(%q) = %q; want %q", lpath, target, f2path)
		}
	} else {
		t.Errorf("os.Readlink(%q): %v", lpath, err)
	}
}

func skipFailTest(t *testing.T, bashPath string) {
	root, deleteTempDir, err := makeTempDir(t)
	if err != nil {
		t.Fatalf("temp directory: %v", err)
	}
	defer deleteTempDir()
	f1path := filepath.Join(root, "foo")
	f2path := filepath.Join(root, "bar")
	f3path := filepath.Join(root, "baz")
	canaryPath := filepath.Join(root, "canary")
	c, err := (&catpogs.Catalog{
		Resources: []*catpogs.Resource{
			{
				ID:      101,
				Comment: "file 1",
				Which:   catalog.Resource_Which_file,
				File:    catpogs.PlainFile(f1path, []byte("foo")),
			},
			{
				ID:      102,
				Deps:    []uint64{101},
				Comment: "file 2",
				Which:   catalog.Resource_Which_file,
				File:    catpogs.PlainFile(f2path, []byte("bar")),
			},
			{
				ID:      103,
				Deps:    []uint64{102},
				Comment: "file 3",
				Which:   catalog.Resource_Which_file,
				File:    catpogs.PlainFile(f3path, []byte("baz")),
			},
			{
				ID:      200,
				Comment: "canary file - not dependent on other files",
				Which:   catalog.Resource_Which_file,
				File:    catpogs.PlainFile(canaryPath, []byte("tweet!")),
			},
		},
	}).ToCapnp()
	if err != nil {
		t.Fatalf("build catalog: %v", err)
	}
	// Create a directory to cause file 1 to fail.
	if err := os.Mkdir(f1path, 0777); err != nil {
		t.Fatal("mkdir:", err)
	}
	_, err = runCatalog("skipFail", bashPath, t, c)
	t.Logf("run catalog: %v", err)
	if err == nil {
		t.Error("run catalog did not return an error")
	}

	if _, err := os.Lstat(f2path); !os.IsNotExist(err) {
		t.Errorf("os.Lstat(%q) = %v; want not exist", f2path, err)
	}
	if _, err := os.Lstat(f3path); !os.IsNotExist(err) {
		t.Errorf("os.Lstat(%q) = %v; want not exist", f3path, err)
	}
	if _, err := os.Lstat(canaryPath); err != nil {
		t.Errorf("os.Lstat(%q) = %v; want nil", canaryPath, err)
	}
}

func execTest(t *testing.T, bashPath string) {
	const msg = "Hello, World!"
	c, err := (&catpogs.Catalog{
		Resources: []*catpogs.Resource{
			{
				ID:      42,
				Comment: "exec",
				Which:   catalog.Resource_Which_exec,
				Exec: &catpogs.Exec{
					Command: &catpogs.Command{
						Which: catalog.Exec_Command_Which_argv,
						Argv:  []string{"/bin/echo", "-n", msg},
					},
				},
			},
		},
	}).ToCapnp()
	if err != nil {
		t.Fatalf("build catalog: %v", err)
	}
	out, err := runCatalog("exec", bashPath, t, c)
	if err != nil {
		t.Errorf("run catalog: %v", err)
	}
	if !bytes.Equal(out, []byte(msg)) {
		t.Errorf("output = %q; want %q", out, msg)
	}
}

func execBashTest(t *testing.T, bashPath string) {
	const msg = "Hello, World!"
	c, err := (&catpogs.Catalog{
		Resources: []*catpogs.Resource{
			{
				ID:      42,
				Comment: "exec",
				Which:   catalog.Resource_Which_exec,
				Exec: &catpogs.Exec{
					Command: &catpogs.Command{
						Which: catalog.Exec_Command_Which_bash,
						Bash:  "/bin/echo -n 'Hello, '\n/bin/echo -n 'World!'\n",
					},
				},
			},
		},
	}).ToCapnp()
	if err != nil {
		t.Fatalf("build catalog: %v", err)
	}
	out, err := runCatalog("execbash", bashPath, t, c)
	if err != nil {
		t.Errorf("run catalog: %v", err)
	}
	if !bytes.Equal(out, []byte(msg)) {
		t.Errorf("output = %q; want %q", out, msg)
	}
}

func execIfDepsChangedTest(t *testing.T, bashPath string) {
	const fileName = "config"
	const fileContent = "Hello"
	const runMessage = "RUNNING"
	makeCatalog := func(root string) (catalog.Catalog, error) {
		fpath := filepath.Join(root, fileName)
		cat, err := (&catpogs.Catalog{
			Resources: []*catpogs.Resource{
				{
					ID:      100,
					Comment: "file",
					Which:   catalog.Resource_Which_file,
					File:    catpogs.PlainFile(fpath, []byte(fileContent)),
				},
				{
					ID:      42,
					Comment: "echo " + runMessage,
					Deps:    []uint64{100},
					Which:   catalog.Resource_Which_exec,
					Exec: &catpogs.Exec{
						Command: &catpogs.Command{
							Which: catalog.Exec_Command_Which_argv,
							Argv:  []string{"/bin/echo", "-n", runMessage},
						},
						Condition: catpogs.ExecCondition{
							Which:         catalog.Exec_condition_Which_ifDepsChanged,
							IfDepsChanged: []uint64{100},
						},
					},
				},
			},
		}).ToCapnp()
		if err != nil {
			return catalog.Catalog{}, fmt.Errorf("build catalog: %v", err)
		}
		return cat, nil
	}
	t.Run("trigger", func(t *testing.T) {
		root, deleteTempDir, err := makeTempDir(t)
		if err != nil {
			t.Fatalf("temp directory: %v", err)
		}
		defer deleteTempDir()
		c, err := makeCatalog(root)
		if err != nil {
			t.Fatal(err)
		}
		out, err := runCatalog("execIfDepsChanged", bashPath, t, c)
		if err != nil {
			t.Errorf("run catalog: %v", err)
		}
		if !bytes.Equal(out, []byte(runMessage)) {
			t.Errorf("output = %q; want %q", out, runMessage)
		}
	})
	t.Run("no-op", func(t *testing.T) {
		root, deleteTempDir, err := makeTempDir(t)
		if err != nil {
			t.Fatalf("temp directory: %v", err)
		}
		defer deleteTempDir()
		fpath := filepath.Join(root, fileName)
		if err := ioutil.WriteFile(fpath, []byte(fileContent), 0666); err != nil {
			t.Fatal("WriteFile:", err)
		}
		c, err := makeCatalog(root)
		if err != nil {
			t.Fatal(err)
		}
		out, err := runCatalog("execIfDepsNotChanged", bashPath, t, c)
		if err != nil {
			t.Errorf("run catalog: %v", err)
		}
		if !bytes.Equal(out, []byte{}) {
			t.Errorf("output = %q; want \"\"", out)
		}
	})
}

const tmpDirEnv = "TEST_TMPDIR"

func runCatalog(scriptName string, bashPath string, log logger, c catalog.Catalog, args ...string) ([]byte, error) {
	sc, err := ioutil.TempFile(os.Getenv(tmpDirEnv), "shlib_testscript_"+scriptName)
	if err != nil {
		return nil, err
	}
	scriptPath := sc.Name()
	if !*keepScripts {
		defer func() {
			if err := os.Remove(scriptPath); err != nil {
				log.Logf("removing temporary script file: %v", err)
			}
		}()
	}
	err = shlib.WriteScript(sc, c)
	cerr := sc.Close()
	if err != nil {
		return nil, err
	}
	if cerr != nil {
		return nil, cerr
	}
	log.Logf("%s -- %s %s", bashPath, scriptPath, strings.Join(args, " "))
	cmd := exec.Command(bashPath, append([]string{"--", scriptPath}, args...)...)
	stdout := new(bytes.Buffer)
	cmd.Stdout = stdout
	stderr := new(bytes.Buffer)
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return stdout.Bytes(), fmt.Errorf("bash failed: %v; stderr:\n%s", err, stderr.Bytes())
	}
	return stdout.Bytes(), nil
}

func makeTempDir(log logger) (path string, done func(), err error) {
	path, err = ioutil.TempDir(os.Getenv(tmpDirEnv), "shlib_testdir")
	if err != nil {
		return "", nil, err
	}
	return path, func() {
		if err := os.RemoveAll(path); err != nil {
			log.Logf("removing temporary directory: %v", err)
		}
	}, nil
}

type logger interface {
	Logf(string, ...interface{})
}
