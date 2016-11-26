package depgraph

import (
	"sort"
	"testing"

	"github.com/zombiezen/mcm/catalog"
	"github.com/zombiezen/mcm/third_party/golang/capnproto"
	"github.com/zombiezen/mcm/third_party/golang/capnproto/pogs"
)

func TestDepgraph(t *testing.T) {
	type Mark struct {
		id      uint64
		fail    bool
		skipped []uint64
	}
	type DummyResource struct {
		ID   uint64   `capnp:"id"`
		Deps []uint64 `capnp:"dependencies"`
	}

	tests := []struct {
		name      string
		skip      bool
		resources []DummyResource
		failNew   bool

		// Marks to apply (in order)
		marks []Mark

		// Conditions to check after applying marks
		ready []uint64
		done  bool
	}{
		{
			name: "empty",
			done: true,
		},
		{
			name: "one resource at start",
			resources: []DummyResource{
				{ID: 42},
			},
			ready: []uint64{42},
		},
		{
			name: "finish one resource",
			resources: []DummyResource{
				{ID: 42},
			},
			marks: []Mark{
				{id: 42},
			},
			done: true,
		},
		{
			name: "fail one resource",
			resources: []DummyResource{
				{ID: 42},
			},
			marks: []Mark{
				{id: 42, fail: true},
			},
			done: true,
		},
		{
			name: "two independent resources",
			resources: []DummyResource{
				{ID: 42},
				{ID: 43},
			},
			ready: []uint64{42, 43},
		},
		{
			name: "two independent resources, finish one",
			resources: []DummyResource{
				{ID: 42},
				{ID: 43},
			},
			marks: []Mark{
				{id: 42},
			},
			ready: []uint64{43},
		},
		{
			name: "two independent resources, fail one",
			resources: []DummyResource{
				{ID: 42},
				{ID: 43},
			},
			marks: []Mark{
				{id: 42, fail: true},
			},
			ready: []uint64{43},
		},
		{
			name: "two independent resources, finish both",
			resources: []DummyResource{
				{ID: 42},
				{ID: 43},
			},
			marks: []Mark{
				{id: 42},
				{id: 43},
			},
			done: true,
		},
		{
			name: "A <- C, B <- C",
			resources: []DummyResource{
				{ID: 10},
				{ID: 20},
				{ID: 30, Deps: []uint64{10, 20}},
			},
			ready: []uint64{10, 20},
		},
		{
			name: "A <- C, B <- C; finish A",
			resources: []DummyResource{
				{ID: 10},
				{ID: 20},
				{ID: 30, Deps: []uint64{10, 20}},
			},
			marks: []Mark{
				{id: 10},
			},
			ready: []uint64{20},
		},
		{
			name: "A <- C, B <- C; finish A and B",
			resources: []DummyResource{
				{ID: 10},
				{ID: 20},
				{ID: 30, Deps: []uint64{10, 20}},
			},
			marks: []Mark{
				{id: 10},
				{id: 20},
			},
			ready: []uint64{30},
		},
		{
			name: "A <- C, B <- C; finish A then B then C",
			resources: []DummyResource{
				{ID: 10},
				{ID: 20},
				{ID: 30, Deps: []uint64{10, 20}},
			},
			marks: []Mark{
				{id: 10},
				{id: 20},
				{id: 30},
			},
			done: true,
		},
		{
			name: "A <- C, B <- C; fail A",
			resources: []DummyResource{
				{ID: 10},
				{ID: 20},
				{ID: 30, Deps: []uint64{10, 20}},
			},
			marks: []Mark{
				{id: 10, fail: true, skipped: []uint64{30}},
			},
			ready: []uint64{20},
		},
		{
			name: "A <- C, B <- C; fail A then mark B",
			resources: []DummyResource{
				{ID: 10},
				{ID: 20},
				{ID: 30, Deps: []uint64{10, 20}},
			},
			marks: []Mark{
				{id: 10, fail: true, skipped: []uint64{30}},
				{id: 20},
			},
			done: true,
		},
		{
			name: "A <- C, B <- C; fail A then fail B",
			resources: []DummyResource{
				{ID: 10},
				{ID: 20},
				{ID: 30, Deps: []uint64{10, 20}},
			},
			marks: []Mark{
				{id: 10, fail: true, skipped: []uint64{30}},
				{id: 20, fail: true},
			},
			done: true,
		},
		{
			name: "A <- B <- C; fail A",
			resources: []DummyResource{
				{ID: 10},
				{ID: 20, Deps: []uint64{10}},
				{ID: 30, Deps: []uint64{20}},
			},
			marks: []Mark{
				{id: 10, fail: true, skipped: []uint64{20, 30}},
			},
			done: true,
		},

		// Cycle tests
		{
			name: "self cycle",
			skip: true,
			resources: []DummyResource{
				{ID: 42, Deps: []uint64{42}},
			},
			failNew: true,
		},
		{
			name: "AB cycle",
			skip: true,
			resources: []DummyResource{
				{ID: 10, Deps: []uint64{20}},
				{ID: 20, Deps: []uint64{10}},
			},
			failNew: true,
		},
		{
			name: "ABC cycle",
			skip: true,
			resources: []DummyResource{
				{ID: 10, Deps: []uint64{20}},
				{ID: 20, Deps: []uint64{30}},
				{ID: 30, Deps: []uint64{10}},
			},
			failNew: true,
		},
		{
			name: "ABC cycle with D",
			skip: true,
			resources: []DummyResource{
				{ID: 10, Deps: []uint64{20}},
				{ID: 20, Deps: []uint64{30}},
				{ID: 30, Deps: []uint64{10}},
				{ID: 40},
			},
			failNew: true,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			if test.skip {
				t.Skip()
			}
			_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
			if err != nil {
				t.Fatal("NewMessage:", err)
			}
			res, err := catalog.NewResource_List(seg, int32(len(test.resources)))
			if err != nil {
				t.Fatal("NewResource_List:", err)
			}
			for i := 0; i < len(test.resources); i++ {
				if err := pogs.Insert(catalog.Resource_TypeID, res.At(i).Struct, &test.resources[i]); err != nil {
					t.Fatalf("insert test.resources[%d]: %v", i, err)
				}
			}

			g, err := New(res)
			if test.failNew {
				if err == nil {
					t.Error("New did not return error")
				}
				return
			}
			if err != nil {
				t.Error("New:", err)
			}

			for _, m := range test.marks {
				if !m.fail {
					t.Logf("g.Mark(%d)", m.id)
					g.Mark(m.id)
					continue
				}
				t.Logf("g.MarkFailure(%d)", m.id)
				skipped := g.MarkFailure(m.id)
				if _, ok := sortSet(skipped); !ok {
					t.Errorf("g.MarkFailure(%d) = %v; do not want duplicates", m.id, skipped)
				}
				if !idSetsEqual(skipped, m.skipped) {
					t.Errorf("g.MarkFailure(%d) = %v; want %v", m.id, skipped, m.skipped)
				}
			}

			if done := g.Done(); done != test.done {
				t.Errorf("g.Done() = %t; want %t", done, test.done)
			}
			ready := g.Ready()
			if _, ok := sortSet(ready); !ok {
				t.Errorf("g.Ready() = %v; do not want duplicates", ready)
			}
			if !idSetsEqual(ready, test.ready) {
				t.Errorf("g.Ready() = %v; want %v", ready, test.ready)
			}
		})
	}
}

func idSetsEqual(a, b []uint64) bool {
	a, _ = sortSet(a)
	b, _ = sortSet(b)
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func sortSet(list []uint64) (set []uint64, ok bool) {
	set = make([]uint64, 0, len(list))
	ok = true
	for _, x := range list {
		i := sort.Search(len(set), func(i int) bool { return set[i] >= x })
		if i < len(set) && set[i] == x {
			ok = false
			continue
		}
		set = append(set, 0)
		copy(set[i+1:], set[i:])
		set[i] = x
	}
	return
}
