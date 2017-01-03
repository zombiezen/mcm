// Copyright 2017 The Minimal Configuration Manager Authors
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

#include "luacat/path.h"

#include <iostream>
#include "gtest/gtest.h"
#include "kj/string.h"

namespace kj {
  inline void PrintTo(const kj::String& s, ::std::ostream* os) {
    os->write(s.begin(), s.size());
  }

  inline void PrintTo(kj::StringPtr s, ::std::ostream* os) {
    os->write(s.begin(), s.size());
  }
}

using mcm::luacat::dirName;
using mcm::luacat::joinPath;

TEST(DirNameTest, NameReturnsCurDir) {
  auto dir = dirName("foo");
  ASSERT_EQ(".", dir);
}

TEST(DirNameTest, ReturnsDirNested1) {
  auto dir = dirName("foo/bar");
  ASSERT_EQ("foo", dir);
}

TEST(DirNameTest, ReturnsDirNested2) {
  auto dir = dirName("foo/bar/baz");
  ASSERT_EQ("foo/bar", dir);
}

TEST(JoinPathTest, OneComponentNop) {
  auto p = joinPath("foo");
  ASSERT_EQ("foo", kj::str(p));
}

TEST(JoinPathTest, TwoComponentsConcatWithSep) {
  auto p = joinPath("foo", "bar");
#if _WIN32
  ASSERT_EQ("foo\\bar", kj::str(p));
#else
  ASSERT_EQ("foo/bar", kj::str(p));
#endif
}

TEST(JoinPathTest, ThreeComponentsConcatWithSep) {
  auto p = joinPath("foo", "bar", "baz");
#if _WIN32
  ASSERT_EQ("foo\\bar\\baz", kj::str(p));
#else
  ASSERT_EQ("foo/bar/baz", kj::str(p));
#endif
}
