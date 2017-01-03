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

#ifndef MCM_LUACAT_TYPES_H_
#define MCM_LUACAT_TYPES_H_
// Custom types for luacat (used in the mcm module).

#include <unistd.h>
#include <stdint.h>

#include "kj/common.h"
#include "kj/debug.h"
#include "kj/memory.h"
#include "kj/string.h"

extern "C" {
#include "lua.h"
}

namespace mcm {

namespace luacat {

class Id {
public:
  inline Id(uint64_t v, kj::StringPtr c) : value(v), comment(kj::heapString(c)) {}
  KJ_DISALLOW_COPY(Id);

  inline uint64_t getValue() const { return value; }
  inline kj::StringPtr getComment() const { return comment; }

private:
  uint64_t value;
  kj::String comment;
};

void pushResourceType(lua_State* state, uint64_t rt);
kj::Maybe<uint64_t> getResourceType(lua_State* state, int index);

void pushId(lua_State* state, kj::Own<const Id> id);
kj::Maybe<const Id&> getId(lua_State* state, int index);

}  // namespace luacat
}  // namespace mcm

#endif  // MCM_LUACAT_TYPES_H_
