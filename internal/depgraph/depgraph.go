// Package depgraph provides a resource dependency graph suitable for scheduling work.
package depgraph

import (
	"errors"
	"fmt"

	"github.com/zombiezen/mcm/catalog"
)

// A Graph schedules work for a DAG of resources.
type Graph struct {
	res   catalog.Resource_List
	deps  map[uint64][]uint64
	index map[uint64]int

	// Mutable state
	ready  []uint64
	queued map[uint64]int
}

// New builds a graph from a list of dependencies or returns an error
// if the dependency information contains inconsistencies.
func New(res catalog.Resource_List) (*Graph, error) {
	n := res.Len()
	g := &Graph{
		res:    res,
		deps:   make(map[uint64][]uint64, n),
		index:  make(map[uint64]int, n),
		queued: make(map[uint64]int, n),
	}
	for i := 0; i < n; i++ {
		id := res.At(i).ID()
		if id == 0 {
			return nil, errors.New("build dependency graph: encountered resource with ID=0")
		}
		g.index[id] = i
		if _, ok := g.deps[id]; !ok {
			g.deps[id] = nil
		}
		deps, err := res.At(i).Dependencies()
		if err != nil {
			return nil, fmt.Errorf("build dependency graph: reading dependency list of resource ID=%d: %v", id, err)
		}
		if ndeps := deps.Len(); ndeps == 0 {
			g.ready = append(g.ready, id)
		} else {
			g.queued[id] = deps.Len()
			for j := 0; j < ndeps; j++ {
				d := deps.At(j)
				g.deps[d] = append(g.deps[d], id)
			}
		}
	}
	for id, out := range g.deps {
		if _, ok := g.index[id]; !ok {
			return nil, fmt.Errorf("build dependency graph: unknown dependency ID %d requested by resource %d", id, out[0])
		}
	}
	// TODO(soon): loop detection
	return g, nil
}

// Ready returns a list of resources that have not been marked and have
// no unmarked dependencies.  This slice is only valid until the next
// mark call.
func (g *Graph) Ready() []uint64 {
	return g.ready
}

// Done returns true if all of the resources in the graph have been marked.
func (g *Graph) Done() bool {
	return len(g.ready)+len(g.queued) == 0
}

// Resource returns the resource with the given ID.
func (g *Graph) Resource(id uint64) catalog.Resource {
	i, ok := g.index[id]
	if !ok {
		return catalog.Resource{}
	}
	return g.res.At(i)
}

// Mark marks a resource as "completed".
func (g *Graph) Mark(id uint64) {
	if !g.pop(id) {
		return
	}
	for _, dep := range g.deps[id] {
		n := g.queued[dep]
		n--
		if n > 0 {
			g.queued[dep] = n
		} else if n == 0 {
			delete(g.queued, dep)
			g.ready = append(g.ready, dep)
		}
	}
}

// Mark marks a resource as "completed with failure" and returns the
// list of resource IDs that depended on this resource, either directly
// or indirectly.  Any resource on the returned list will never appear
// in the ready list.
func (g *Graph) MarkFailure(id uint64) []uint64 {
	if !g.pop(id) {
		return nil
	}
	var aborted []uint64
	visited := func(id uint64) bool {
		for _, x := range aborted {
			if x == id {
				return true
			}
		}
		return false
	}
	var stk []uint64
	for {
		for _, dep := range g.deps[id] {
			if g.queued[dep] != 0 && !visited(dep) {
				stk = append(stk, dep)
			}
		}
		if len(stk) == 0 {
			break
		}
		end := len(stk) - 1
		stk, id = stk[:end], stk[end]
		aborted = append(aborted, id)
		delete(g.queued, id)
	}
	return aborted
}

func (g *Graph) pop(id uint64) bool {
	i := -1
	for ii, r := range g.ready {
		if r == id {
			i = ii
			break
		}
	}
	if i == -1 {
		return false
	}
	g.ready = append(g.ready[:i], g.ready[i+1:]...)
	return true
}
