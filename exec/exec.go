package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/zombiezen/mcm/catalog"
	"github.com/zombiezen/mcm/exec/execlib"
	"github.com/zombiezen/mcm/third_party/golang/capnproto"
)

func init() {
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [CATALOG]:\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	log := new(logger)
	o := &realOS{log: log}
	app := &execlib.Applier{
		OS:  o,
		Log: log,
	}
	flag.BoolVar(&app.Unconditional, "skip-conditions", false, "whether to skip conditions")
	flag.BoolVar(&o.simulate, "n", false, "dry-run")
	flag.BoolVar(&log.quiet, "q", false, "suppress info messages and failure output")
	flag.BoolVar(&o.logCommands, "s", false, "show commands run in the log")
	flag.Parse()

	ctx := context.Background()
	var cat catalog.Catalog
	switch flag.NArg() {
	case 0:
		var err error
		cat, err = readCatalog(os.Stdin)
		if err != nil {
			log.Fatal(ctx, err)
		}
	case 1:
		// TODO(someday): read segments lazily
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			log.Fatal(ctx, err)
		}
		cat, err = readCatalog(f)
		if err != nil {
			log.Fatal(ctx, err)
		}
		if err = f.Close(); err != nil {
			log.Error(ctx, err)
		}
	default:
		usage()
		os.Exit(2)
	}

	if err := app.Apply(ctx, cat); err != nil {
		log.Fatal(ctx, err)
	}
}

type realOS struct {
	simulate    bool
	logCommands bool
	log         *logger
}

func (o *realOS) Lstat(path string) (os.FileInfo, error) {
	// Allow stat even when simulated.
	return os.Lstat(path)
}

func (o *realOS) WriteFile(path string, content []byte, mode os.FileMode) error {
	if o.simulate {
		return nil
	}
	return ioutil.WriteFile(path, content, mode)
}

func (o *realOS) Mkdir(path string, mode os.FileMode) error {
	if o.simulate {
		return nil
	}
	return os.Mkdir(path, mode)
}

func (o *realOS) Remove(path string) error {
	if o.simulate {
		return nil
	}
	return os.Remove(path)
}

func (o *realOS) Run(ctx context.Context, cmd *exec.Cmd) (output []byte, err error) {
	if o.logCommands {
		o.log.Infof(ctx, "exec %s", strings.Join(cmd.Args, " "))
	}
	if o.simulate {
		return nil, nil
	}
	return cmd.CombinedOutput()
}

type logger struct {
	quiet bool
	mu    sync.Mutex
}

func (l *logger) Infof(ctx context.Context, format string, args ...interface{}) {
	if l.quiet {
		return
	}
	now := time.Now()
	var line bytes.Buffer
	writeLogHead(&line, "INFO", now)
	fmt.Fprintf(&line, format, args...)
	if b := line.Bytes(); b[len(b)-1] != '\n' {
		line.WriteByte('\n')
	}
	defer l.mu.Unlock()
	l.mu.Lock()
	os.Stderr.Write(line.Bytes())
}

func (l *logger) Error(ctx context.Context, err error) {
	now := time.Now()
	var line bytes.Buffer
	writeLogHead(&line, "ERROR", now)
	line.WriteString(err.Error())
	if b := line.Bytes(); b[len(b)-1] != '\n' {
		line.WriteByte('\n')
	}

	var output []byte
	if !l.quiet {
		if err, ok := err.(*execlib.Error); ok && len(err.Output) > 0 {
			output = err.Output
			if n := len(output); output[n-1] == '\n' {
				new := make([]byte, n+1)
				copy(new, output)
				new[n] = '\n'
				output = new
			}
			output = err.Output
			if err.Output[len(err.Output)-1] != '\n' {
				line.WriteByte('\n')
			}
		}
	}

	defer l.mu.Unlock()
	l.mu.Lock()
	os.Stderr.Write(line.Bytes())
	if len(output) > 0 {
		os.Stderr.Write(output)
	}
}

func writeLogHead(buf *bytes.Buffer, severity string, now time.Time) {
	buf.WriteString("mcm-exec: ")
	buf.WriteString(now.Format("2006-01-02T15:04:05"))
	fmt.Fprintf(buf, " %5s: ", severity)
}

func (l *logger) Fatal(ctx context.Context, err error) {
	l.Error(ctx, err)
	os.Exit(1)
}

func readCatalog(r io.Reader) (catalog.Catalog, error) {
	msg, err := capnp.NewDecoder(r).Decode()
	if err != nil {
		return catalog.Catalog{}, fmt.Errorf("read catalog: %v", err)
	}
	c, err := catalog.ReadRootCatalog(msg)
	if err != nil {
		return catalog.Catalog{}, fmt.Errorf("read catalog: %v", err)
	}
	return c, nil
}
