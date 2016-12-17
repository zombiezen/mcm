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

package catpogs

import (
	"encoding/hex"
	"testing"

	"github.com/zombiezen/mcm/catalog"
)

func TestNoContentFile(t *testing.T) {
	c, err := (&Catalog{
		Resources: []*Resource{
			{
				ID:      42,
				Comment: "file",
				Which:   catalog.Resource_Which_file,
				File:    PlainFile("/foo", nil),
			},
		},
	}).ToCapnp()
	if err != nil {
		t.Fatalf("ToCapnp: %v", err)
	}
	r, err := c.Resources()
	if err != nil {
		t.Fatalf("resources: %v", err)
	}
	rr := r.At(0)
	if rr.Which() != catalog.Resource_Which_file {
		t.Fatalf("resources[0] is %v; want file", rr.Which())
	}
	f, err := rr.File()
	if err != nil {
		t.Fatalf("resources[0].file: %v", err)
	}
	if f.Which() != catalog.File_Which_plain {
		t.Fatalf("resources[0].file is %v; want plain", f.Which())
	}
	if f.Plain().HasContent() {
		data, _ := c.Segment().Message().Marshal()
		t.Log("dump:\n" + hex.Dump(data))
		t.Fatal("resources[0].file.content is not null")
	}
}
