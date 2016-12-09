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

func TestIntegration(t *testing.T) {
	bashPath, err := exec.LookPath("bash")
	if err != nil {
		t.Skipf("Can't find bash: %v", err)
	}
	t.Logf("using %s for bash", bashPath)
	t.Run("file", func(t *testing.T) { fileIntegrationTest(t, bashPath) })
}

func fileIntegrationTest(t *testing.T, bashPath string) {
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
	_, err = runCatalog(bashPath, t, c)
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

const tmpDirEnv = "TEST_TMPDIR"

func runCatalog(bashPath string, log logger, c catalog.Catalog, args ...string) ([]byte, error) {
	sc, err := ioutil.TempFile(os.Getenv(tmpDirEnv), "shlib_testscript")
	if err != nil {
		return nil, err
	}
	scriptPath := sc.Name()
	defer func() {
		if err := os.Remove(scriptPath); err != nil {
			log.Logf("removing temporary script file: %v", err)
		}
	}()
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
