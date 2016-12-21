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
  ]
);
