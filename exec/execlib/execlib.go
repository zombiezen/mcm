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
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/zombiezen/mcm/catalog"
	"github.com/zombiezen/mcm/internal/depgraph"
	"github.com/zombiezen/mcm/internal/system"
)

// DefaultBashPath is the path used if Applier.Bash is empty.
const DefaultBashPath = "/bin/bash"

// An Applier applies a catalog to a system.
type Applier struct {
	System system.System
	Log    Logger

	// Bash is the path to the bash executable.
	// If it's empty, then Apply uses DefaultBashPath.
	Bash string

	// ConcurrentJobs is the number of resources to apply simultaneously.
	// If non-positive, then it assumes 1.
	ConcurrentJobs int

	// TODO(someday): pass this down as an argument
	userLookup userLookupCache
}

// Logger collects execution messages from an Applier.  A Logger must be
// safe to call from multiple goroutines.
type Logger interface {
	Infof(ctx context.Context, format string, args ...interface{})
	Error(ctx context.Context, err error)
}

func (app *Applier) Apply(ctx context.Context, c catalog.Catalog) error {
	res, _ := c.Resources()
	g, err := depgraph.New(res)
	if err != nil {
		return toError(err)
	}
	app.userLookup = userLookupCache{lookup: app.System}
	if err = app.applyCatalog(ctx, g); err != nil {
		return toError(err)
	}
	return nil
}

type applyState struct {
	graph            *depgraph.Graph
	hasFailures      bool
	changedResources map[uint64]bool
}

func (app *Applier) applyCatalog(ctx context.Context, g *depgraph.Graph) error {
	njobs := app.ConcurrentJobs
	if njobs < 1 {
		njobs = 1
	}
	ch, results, done := app.startWorkers(ctx, njobs)
	defer done()

	state := &applyState{
		graph:            g,
		changedResources: make(map[uint64]bool),
	}
	working := make(workingSet, njobs)
	var nextArgs applyArgs
	for !g.Done() {
		if working.hasIdle() && !nextArgs.resource.IsValid() {
			// Find next work, if any.
			ready := g.Ready()
			if len(ready) == 0 {
				return errors.New("graph not done, but has nothing to do")
			}
			if id := working.next(ready); id != 0 {
				res := g.Resource(id)
				nextArgs = applyArgs{
					resource:   res,
					depChanged: mapChangedDeps(state.changedResources, res),
				}
			}
		}
		if !nextArgs.resource.IsValid() {
			select {
			case r := <-results:
				working.remove(r.id)
				app.update(ctx, state, r)
			case <-ctx.Done():
				return ctx.Err()
			}
			continue
		}
		select {
		case ch <- nextArgs:
			working.add(nextArgs.resource.ID())
			nextArgs = applyArgs{}
		case r := <-results:
			working.remove(r.id)
			app.update(ctx, state, r)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	close(ch)
	if state.hasFailures {
		return errors.New("not all resources applied cleanly")
	}
	return nil
}

func (app *Applier) update(ctx context.Context, state *applyState, r applyResult) {
	if r.err != nil {
		state.hasFailures = true
		res := state.graph.Resource(r.id)
		app.Log.Error(ctx, errorWithResource(res, r.err))
		skipped := state.graph.MarkFailure(r.id)
		if len(skipped) == 0 {
			return
		}
		skipnames := make([]string, len(skipped))
		for i := range skipnames {
			skipnames[i] = formatResource(state.graph.Resource(skipped[i]))
		}
		app.Log.Infof(ctx, "skipping due to failure of %s: %s", formatResource(res), strings.Join(skipnames, ", "))
		return
	}
	state.graph.Mark(r.id)
	state.changedResources[r.id] = r.changed
}

func mapChangedDeps(all map[uint64]bool, r catalog.Resource) map[uint64]bool {
	deps, _ := r.Dependencies()
	n := deps.Len()
	m := make(map[uint64]bool, n)
	for i := 0; i < n; i++ {
		d := deps.At(i)
		m[d] = all[d]
	}
	return m
}

// workingSet is the list of resource IDs being processed at a point in time.
type workingSet []uint64

// hasIdle reports whether there are idle workers.
func (ws workingSet) hasIdle() bool {
	return ws.find(0) != -1
}

func (ws workingSet) add(id uint64) {
	i := ws.find(0)
	if i == -1 {
		panic("workingSet.add on full set")
	}
	ws[i] = id
}

func (ws workingSet) remove(id uint64) {
	i := ws.find(id)
	if i == -1 {
		panic("workingSet.remove could not find ID")
	}
	ws[i] = 0
}

// find returns the index of an ID in the set or -1 if not found.
func (ws workingSet) find(id uint64) int {
	for i := range ws {
		if ws[i] == id {
			return i
		}
	}
	return -1
}

// next returns a resource ID that is in ready but not in ws or zero if ws is a superset of ready.
func (ws workingSet) next(ready []uint64) uint64 {
	// While this is technically O(len(ws) * len(ready)),
	// len(ws) is constant over the course of an Apply.
	for _, id := range ready {
		if ws.find(id) == -1 {
			return id
		}
	}
	return 0
}

type applyArgs struct {
	resource   catalog.Resource
	depChanged map[uint64]bool
}

type applyResult struct {
	id      uint64
	changed bool
	err     error
}

func (app *Applier) startWorkers(ctx context.Context, n int) (chan<- applyArgs, <-chan applyResult, func()) {
	ch := make(chan applyArgs)
	results := make(chan applyResult)
	workCtx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			app.worker(workCtx, results, ch)
			wg.Done()
		}()
	}
	return ch, results, func() {
		cancel()
		wg.Wait()
	}
}

func (app *Applier) worker(ctx context.Context, results chan<- applyResult, ch <-chan applyArgs) {
	for {
		select {
		case args, ok := <-ch:
			if !ok {
				return
			}
			app.Log.Infof(ctx, "applying: %s", formatResource(args.resource))
			changed, err := app.applyResource(ctx, args.resource, args.depChanged)
			r := applyResult{
				id:      args.resource.ID(),
				changed: changed,
				err:     err,
			}
			select {
			case results <- r:
			case <-ctx.Done():
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// TODO(someday): ensure lookups are single-flight

type userLookupCache struct {
	lookup system.UserLookup

	mu     sync.RWMutex
	users  map[string]system.UID
	groups map[string]system.GID
}

func (c *userLookupCache) LookupUser(name string) (system.UID, error) {
	c.mu.RLock()
	uid, ok := c.users[name]
	c.mu.RUnlock()
	if ok {
		return uid, nil
	}

	uid, err := c.lookup.LookupUser(name)
	if err != nil {
		return uid, err
	}
	c.mu.Lock()
	if c.users == nil {
		c.users = make(map[string]system.UID)
	}
	c.users[name] = uid
	c.mu.Unlock()
	return uid, nil
}

func (c *userLookupCache) LookupGroup(name string) (system.GID, error) {
	c.mu.RLock()
	gid, ok := c.groups[name]
	c.mu.RUnlock()
	if ok {
		return gid, nil
	}
	gid, err := c.lookup.LookupGroup(name)
	if err != nil {
		return gid, err
	}
	c.mu.Lock()
	if c.groups == nil {
		c.groups = make(map[string]system.GID)
	}
	c.groups[name] = gid
	c.mu.Unlock()
	return gid, nil
}

func formatResource(r catalog.Resource) string {
	c, _ := r.Comment()
	if c == "" {
		return fmt.Sprintf("id=%d", r.ID())
	}
	return fmt.Sprintf("%s (id=%d)", c, r.ID())
}
