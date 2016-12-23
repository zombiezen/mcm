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
// Conversions between Lua built-in types and C++ (KJ) types.

#include <unistd.h>

#include "kj/array.h"
#include "kj/exception.h"
#include "kj/io.h"
#include "kj/string.h"
#include "capnp/dynamic.h"

extern "C" {
#include "lua.h"
}

namespace mcm {

namespace luacat {

const kj::ArrayPtr<const kj::byte> luaBytePtr(lua_State* state, int index);
// Converts the Lua value at the given index to a byte array.
// The memory is owned by Lua, so the caller must keep the Lua value on
// the stack while the return value is live.

const kj::StringPtr luaStringPtr(lua_State* state, int index);
// Converts the Lua value at the given index to a string.
// The memory is owned by Lua, so the caller must keep the Lua value on
// the stack while the return value is live.

int luaLoad(lua_State* state, kj::StringPtr name, kj::InputStream& stream);

inline void pushLua(lua_State* state, const kj::StringPtr s) {
  // Push a string onto the Lua stack.
  lua_pushlstring(state, s.cStr(), s.size());
}

void pushLua(lua_State* state, kj::Exception& e);
// Push a description of e onto the Lua stack.

void copyStruct(lua_State* state, capnp::DynamicStruct::Builder builder);
// Converts the Lua value at the top of the stack into a Cap'n Proto struct.
// Throws kj::Exception on input validation error.

void copyList(lua_State* state, capnp::DynamicList::Builder builder);
// Converts the Lua value at the top of the stack into a Cap'n Proto list.
// Throws kj::Exception on input validation error.

}  // namespace luacat
}  // namespace mcm

#endif  // MCM_LUACAT_TYPES_H_
