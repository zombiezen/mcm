# Copyright 2016 The Minimal Configuration Manager Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

using Cxx = import "/third_party/capnproto/c++/src/capnp/c++.capnp";
using import "/catalog.capnp".Catalog;

@0xa5aa8d2c7f415790;
$Cxx.namespace("mcm::luacat");
# Definition of test suite.

struct TestSuite {
  tests @0 :List(TestCase);
}

struct TestCase {
  name @0 :Text;
  script @1 :Text;
  
  expected :group {
    output @2 :Text;

    union {
      success @3 :Void;
      error @4 :Void;
      catalog @5 :Catalog;
    }
  }
}

struct GenericValue {
  union {
    void @0 :Void;
    bool @1 :Bool;
    int8 @2 :Int8;
    int16 @3 :Int16;
    int32 @4 :Int32;
    int64 @5 :Int64;
    uint8 @6 :UInt8;
    uint16 @7 :UInt16;
    uint32 @8 :UInt32;
    uint64 @9 :UInt64;
    float32 @10 :Float32;
    float64 @11 :Float64;
    text @12 :Text;
    data @13 :Data;

    structList @14 :List(GenericValue);
    boolList @17 :List(Bool);
    listList @18 :List(List(Int16));
    enumList @19 :List(Subject);
    uint64List @20 :List(UInt64);

    enum @15 :Subject;
    struct @16 :GenericValue;
  }
}

enum Subject {
  this @0;
  that @1;
  other @2;
}

const testSuite :TestSuite = (
  tests = [
    (
      name = "empty script",
      script = "",
      expected = (success = void),
    ),
    (
      name = "hello world",
      script = "print(\"Hello, World!\")\n",
      expected = (output = "Hello, World!\n"),
    ),
    (
      name = "mcm global exists",
      script = "print(mcm ~= nil)\n",
      expected = (output = "true\n"),
    ),
    (
      name = "file resource",
      script = embed "testdata/file.lua",
      expected = (
        catalog = (
          resources = [
            (
              id = 0x20e102f0f9e2b11d,
              comment = "hello",
              file = (
                path = "/etc/hello.txt",
                plain = (
                  content = "Hello, World!\n",
                ),
              ),
            ),
          ],
        ),
      ),
    ),
    (
      name = "deps changed",
      script = embed "testdata/depschanged.lua",
      expected = (
        catalog = (
          resources = [
            (
              id = 0xd96f419065c49db1,
              comment = "xyzzy!",
              file = (
                path = "/etc/motd",
                plain = (),
              ),
            ),
            (
              id = 0x3d784cfc26097123,
              comment = "apt-get update",
              dependencies = [0xd96f419065c49db1],
              exec = (
                condition = (
                  ifDepsChanged = [0xd96f419065c49db1],
                ),
                command = (
                  argv = ["/usr/bin/apt-get", "update"],
                ),
              ),
            ),
          ],
        ),
      ),
    ),
  ]
);
