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

void copyStruct(lua_State* state, capnp::DynamicStruct::Builder builder);
// Converts the Lua value at the top of the stack into a Cap'n Proto struct.

void copyList(lua_State* state, capnp::DynamicList::Builder builder);
// Converts the Lua value at the top of the stack into a Cap'n Proto list.

}  // namespace luacat
}  // namespace mcm

#endif  // MCM_LUACAT_TYPES_H_
