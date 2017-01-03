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

#ifndef MCM_LUACAT_LIB_H_
#define MCM_LUACAT_LIB_H_
// mcm Lua module.

#include "kj/common.h"
#include "kj/vector.h"
#include "capnp/message.h"

extern "C" {
#include "lua.h"
}

#include "catalog.capnp.h"

namespace mcm {

namespace luacat {

class LibState {
  // The mutable state of the mcm Lua module.
public:
  LibState() {}
  KJ_DISALLOW_COPY(LibState);

  Resource::Builder newResource();
  inline kj::ArrayPtr<capnp::Orphan<Resource>> getResources() { return resources.asPtr(); }
private:
  capnp::MallocMessageBuilder scratch;
  kj::Vector<capnp::Orphan<Resource>> resources;
};

void openlib(lua_State* state, LibState& lib);
// Loads the "mcm" module and leaves it at the top of the stack.

}  // namespace luacat
}  // namespace mcm

#endif  // MCM_LUACAT_LIB_H_
