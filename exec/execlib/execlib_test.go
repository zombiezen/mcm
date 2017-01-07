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

package execlib

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"testing"

	"github.com/zombiezen/mcm/catalog"
	. "github.com/zombiezen/mcm/exec/execlib"
	"github.com/zombiezen/mcm/internal/applytests"
	"github.com/zombiezen/mcm/internal/catpogs"
	"github.com/zombiezen/mcm/internal/system"
	"github.com/zombiezen/mcm/internal/system/fakesystem"
)

func TestApplier(t *testing.T) {
	applytests.Run(t, new(fixtureFactory).newFixture)
}

func TestApplier2Jobs(t *testing.T) {
	applytests.Run(t, (&fixtureFactory{concurrentJobs: 2}).newFixture)
}

func TestExecBash(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const scriptOutput = "Hello, World!"
	const script = "/bin/echo -n '" + scriptOutput + "'\n"
	cat, err := (&catpogs.Catalog{
		Resources: []*catpogs.Resource{
			{
				ID:      42,
				Comment: "exec",
				Which:   catalog.Resource_Which_exec,
				Exec: &catpogs.Exec{
					Command: &catpogs.Command{
						Which: catalog.Exec_Command_Which_bash,
						Bash:  script,
					},
				},
			},
		},
	}).ToCapnp()
	if err != nil {
		t.Fatal("catpogs.Catalog.ToCapnp():", err)
	}
	binPath := filepath.Join(fakesystem.Root, "xyzzybin")
	bashPath := filepath.Join(binPath, "bash")
	sys := new(fakesystem.System)
	if err := sys.Mkdir(ctx, binPath, 0777); err != nil {
		t.Fatalf("mkdir %s: %v", binPath, err)
	}
	scriptBuf := new(bytes.Buffer)
	err = sys.Mkprogram(bashPath, func(ctx context.Context, pc *fakesystem.ProgramContext) int {
		if len(pc.Args) != 1 || pc.Args[0] != bashPath {
			fmt.Fprintf(pc.Output, "arguments = %v; want [%s]\n", pc.Args, bashPath)
			return 1
		}
		if _, err := io.Copy(scriptBuf, pc.Input); err != nil {
			fmt.Fprintln(pc.Output, err)
			return 1
		}
		if _, err := io.WriteString(pc.Output, scriptOutput); err != nil {
			fmt.Fprintln(pc.Output, err)
			return 1
		}
		return 0
	})
	if err != nil {
		t.Fatal("Mkprogram:", err)
	}

	app := &Applier{
		System: sys,
		Log:    testLogger{t: t},
		Bash:   bashPath,
	}
	err = app.Apply(ctx, cat)
	if err != nil {
		t.Error("Apply:", err)
	}

	if got := scriptBuf.String(); got != script {
		t.Errorf("script = %q; want %q", got, script)
	}
}

type fixtureFactory struct {
	concurrentJobs int
}

type fixture struct {
	sys            *fakesystem.System
	log            applytests.Logger
	info           *applytests.SystemInfo
	concurrentJobs int
}

func (ff *fixtureFactory) newFixture(ctx context.Context, log applytests.Logger, name string) (applytests.Fixture, error) {
	sys := new(fakesystem.System)
	binPath := filepath.Join(fakesystem.Root, "mybin")
	info := &applytests.SystemInfo{
		Root:      filepath.Join(fakesystem.Root, "subdir"),
		TruePath:  filepath.Join(binPath, "true"),
		FalsePath: filepath.Join(binPath, "false"),
		TouchPath: filepath.Join(binPath, "touch"),
	}
	if err := sys.Mkdir(ctx, binPath, 0755); err != nil {
		return nil, err
	}
	if err := sys.Mkdir(ctx, info.Root, 0755); err != nil {
		return nil, err
	}
	err := sys.Mkprogram(info.TruePath, func(ctx context.Context, pc *fakesystem.ProgramContext) int {
		return 0
	})
	if err != nil {
		return nil, err
	}
	err = sys.Mkprogram(info.FalsePath, func(ctx context.Context, pc *fakesystem.ProgramContext) int {
		return 1
	})
	if err != nil {
		return nil, err
	}
	err = sys.Mkprogram(info.TouchPath, func(ctx context.Context, pc *fakesystem.ProgramContext) int {
		if len(pc.Args) != 2 {
			fmt.Fprintln(pc.Output, "usage: touch FILE")
			return 1
		}
		w, err := sys.CreateFile(ctx, pc.Args[1], 0666)
		if err != nil {
			fmt.Fprintln(pc.Output, "touch:", err)
			return 1
		}
		if err := w.Close(); err != nil {
			fmt.Fprintln(pc.Output, "touch:", err)
			return 1
		}
		return 0
	})
	if err != nil {
		return nil, err
	}
	return &fixture{
		sys:            sys,
		log:            log,
		info:           info,
		concurrentJobs: ff.concurrentJobs,
	}, nil
}

func (f *fixture) Apply(ctx context.Context, c catalog.Catalog) error {
	app := &Applier{
		System:         f.sys,
		Log:            testLogger{t: f.log},
		ConcurrentJobs: f.concurrentJobs,
	}
	return app.Apply(ctx, c)
}

func (f *fixture) System() system.System {
	return f.sys
}

func (f *fixture) SystemInfo() *applytests.SystemInfo {
	return f.info
}

func (f *fixture) Close() error {
	return nil
}

type testLogger struct {
	t applytests.Logger
}

func (tl testLogger) Infof(ctx context.Context, format string, args ...interface{}) {
	tl.t.Logf("applier info: "+format, args...)
}

func (tl testLogger) Error(ctx context.Context, err error) {
	tl.t.Logf("applier error: %v", err)
}
