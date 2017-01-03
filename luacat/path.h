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

#ifndef MCM_LUACAT_PATH_H_
#define MCM_LUACAT_PATH_H_
// Path manipulation routines.

#include "kj/common.h"
#include "kj/string.h"
#include "kj/string-tree.h"

namespace mcm {

namespace luacat {

kj::String dirName(kj::StringPtr path);

namespace _ {
#if _WIN32
  const char pathSep = '\\';
#else
  const char pathSep = '/';
#endif

  inline kj::StringTree joinPath(kj::StringTree&& tree) { return kj::mv(tree); }

  template <typename First, typename... Rest>
  kj::StringTree joinPath(kj::StringTree&& tree, const First& first, Rest&&... rest) {
    return joinPath(kj::strTree(kj::mv(tree), pathSep, first), kj::fwd<Rest>(rest)...);
  }
}  // namespace _

template <typename First, typename... Rest>
kj::StringTree joinPath(const First& first, Rest&&... rest) {
  // Concatenate a bunch of stringifiable components into a single path,
  // separated by the OS's path separator.

  return _::joinPath(kj::strTree(first), kj::fwd<Rest>(rest)...);
}

}  // namespace luacat
}  // namespace mcm

#endif  // MCM_LUACAT_PATH_H_
