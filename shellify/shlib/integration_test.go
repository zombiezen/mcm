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
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/zombiezen/mcm/catalog"
	"github.com/zombiezen/mcm/internal/applytests"
	"github.com/zombiezen/mcm/internal/catpogs"
	"github.com/zombiezen/mcm/internal/system"
	"github.com/zombiezen/mcm/shellify/shlib"
)

var keepScripts = flag.Bool("keep_scripts", false, "do not remove generated scripts from temporary directory")

const tmpDirEnv = "TEST_TMPDIR"

func TestIntegration(t *testing.T) {
	bashPath, err := exec.LookPath("bash")
	if err != nil {
		t.Skipf("Can't find bash: %v", err)
	}
	t.Logf("using %s for bash", bashPath)
	applytests.Run(t, (&fixtureFactory{bashPath: bashPath}).newFixture)
}

func TestExecBash(t *testing.T) {
	bashPath, err := exec.LookPath("bash")
	if err != nil {
		t.Skipf("Can't find bash: %v", err)
	}
	t.Logf("using %s for bash", bashPath)

	ctx, cancel := context.WithCancel(context.Background())
	f, err := (&fixtureFactory{bashPath: bashPath}).newFixture(ctx, t, "execbash")
	if err != nil {
		cancel()
		t.Fatal("fixture:", err)
	}
	defer func() {
		cancel()
		if err := f.Close(); err != nil {
			t.Error("fixture close:", err)
		}
	}()

	info := f.SystemInfo()
	fpath := filepath.Join(info.Root, "canary")
	c, err := (&catpogs.Catalog{
		Resources: []*catpogs.Resource{
			{
				ID:      42,
				Comment: "exec",
				Which:   catalog.Resource_Which_exec,
				Exec: &catpogs.Exec{
					Command: &catpogs.Command{
						Which: catalog.Exec_Command_Which_bash,
						Bash:  info.TouchPath + " '" + fpath + "'\n",
					},
				},
			},
		},
	}).ToCapnp()
	if err != nil {
		t.Fatalf("build catalog: %v", err)
	}
	err = f.Apply(ctx, c)
	if err != nil {
		t.Errorf("run catalog: %v", err)
	}
	if _, err := os.Lstat(fpath); err != nil {
		t.Errorf("checking for %q: %v", fpath, err)
	}
}

type fixtureFactory struct {
	bashPath string
}

func (ff *fixtureFactory) newFixture(ctx context.Context, log applytests.Logger, name string) (applytests.Fixture, error) {
	f := &fixture{
		name:     name,
		log:      log,
		bashPath: ff.bashPath,
	}
	var err error
	f.root, err = ioutil.TempDir(os.Getenv(tmpDirEnv), "shlib_testdir")
	if err != nil {
		return nil, err
	}
	return f, nil
}

type fixture struct {
	name     string
	log      applytests.Logger
	bashPath string

	root string
}

func (f *fixture) System() system.System {
	return system.Local{}
}

func (f *fixture) SystemInfo() *applytests.SystemInfo {
	return &applytests.SystemInfo{
		Root:      f.root,
		TruePath:  "/bin/true",
		FalsePath: "/bin/false",
		TouchPath: "/usr/bin/touch",
	}
}

func (f *fixture) Apply(ctx context.Context, c catalog.Catalog) error {
	sc, err := ioutil.TempFile(os.Getenv(tmpDirEnv), "shlib_testscript_"+f.name)
	if err != nil {
		return err
	}
	scriptPath := sc.Name()
	if !*keepScripts {
		defer func() {
			if err := os.Remove(scriptPath); err != nil {
				f.log.Logf("removing temporary script file: %v", err)
			}
		}()
	}
	err = shlib.WriteScript(sc, c)
	cerr := sc.Close()
	if err != nil {
		return err
	}
	if cerr != nil {
		return cerr
	}
	f.log.Logf("%s -- %s", f.bashPath, scriptPath)
	cmd := exec.Command(f.bashPath, []string{"--", scriptPath}...)
	stderr := new(bytes.Buffer)
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("bash failed: %v; stderr:\n%s", err, stderr.Bytes())
	}
	return nil
}

func (f *fixture) Close() error {
	if err := os.RemoveAll(f.root); err != nil {
		return fmt.Errorf("removing temporary directory: %v", err)
	}
	return nil
}
