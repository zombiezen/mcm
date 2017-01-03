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

namespace {
  template <typename T, typename E>
  size_t count(T iter, E elem) {
    size_t n = 0;
    for (E x: iter) {
      if (x == elem) {
        n++;
      }
    }
    return n;
  }
}  // namespace

namespace mcm {

namespace luacat {

kj::String dirName(kj::StringPtr path) {
  KJ_IF_MAYBE(slashPos, path.findLast(_::pathSep)) {
    return kj::heapString(path.slice(0, *slashPos));
  } else {
    return kj::heapString(".");
  }
}

kj::Array<kj::ArrayPtr<const char>> splitStr(kj::StringPtr s, char delim) {
  auto parts = kj::heapArray<kj::ArrayPtr<const char>>(count(s, delim) + 1);
  size_t n = 0;
  size_t last = 0;
  for (size_t i = 0; i < s.size(); i++) {
    if (s[i] != delim) {
      continue;
    }
    parts[n++] = s.slice(last, i);
    last = i + 1;
  }
  parts[n] = s.slice(last);
  return parts;
}

}  // namespace luacat
}  // namespace mcm
