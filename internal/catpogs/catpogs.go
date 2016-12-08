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

// Package catpogs provides Go struct equivalents of the Cap'n Proto
// catalog.  These are primarily intended for constructing test inputs.
package catpogs

import (
	"github.com/zombiezen/mcm/catalog"
	"github.com/zombiezen/mcm/third_party/golang/capnproto"
	"github.com/zombiezen/mcm/third_party/golang/capnproto/pogs"
)

type Catalog struct {
	Resources []*Resource
}

func (c *Catalog) ToCapnp() (catalog.Catalog, error) {
	_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return catalog.Catalog{}, err
	}
	root, err := catalog.NewRootCatalog(seg)
	if err != nil {
		return catalog.Catalog{}, err
	}
	err = pogs.Insert(catalog.Catalog_TypeID, root.Struct, c)
	return root, err
}

type Resource struct {
	ID      uint64 `capnp:"id"`
	Comment string
	Deps    []uint64 `capnp:"dependencies"`

	Which catalog.Resource_Which
	File  *File
	Exec  *Exec
}

type File struct {
	Path string

	Which catalog.File_Which
	Plain struct {
		Content []byte
	}
	Directory struct{}
	Symlink   struct {
		Target string
	}
}

func PlainFile(path string, content []byte) *File {
	f := &File{
		Path:  path,
		Which: catalog.File_Which_plain,
	}
	f.Plain.Content = content
	return f
}

func SymlinkFile(oldname, newname string) *File {
	f := &File{
		Path:  newname,
		Which: catalog.File_Which_symlink,
	}
	f.Symlink.Target = oldname
	return f
}

type Exec struct {
	Command   *Command
	Condition ExecCondition
}

type ExecCondition struct {
	Which         catalog.Exec_condition_Which
	OnlyIf        *Command
	Unless        *Command
	FileAbsent    string
	IfDepsChanged []uint64
}

type Command struct {
	Which catalog.Exec_Command_Which
	Argv  []string
	Bash  string

	Env []EnvVar `capnp:"environment"`
	Dir string   `capnp:"workingDirectory"`
}

type EnvVar struct {
	Name, Value string
}
