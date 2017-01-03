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

#include "luacat/types.h"

#include "lua.hpp"

namespace mcm {

namespace luacat {

namespace {
  const char* resourceTypeKey = "mcm resourcetype";
  const char* idKey = "mcm id";

  template<typename T>
  T& newUserData(lua_State* state) {
    auto p = reinterpret_cast<T*>(lua_newuserdata(state, sizeof(T)));
    kj::ctor(*p);
    return *p;
  }

  template<typename T>
  kj::Maybe<T&> testUserData(lua_State* state, int index, const char* tname) {
    void* p = lua_touserdata(state, index);
    if (p == nullptr) {  // value is a userdata?
      return nullptr;  // value is not a userdata with a metatable
    }
    if (lua_getmetatable(state, index)) {  // does it have a metatable?
      if (luaL_getmetatable(state, tname) == LUA_TNIL) {  // get correct metatable
        p = nullptr;  // metatable not set up yet; can't have the metatable
      } else if (!lua_rawequal(state, -1, -2)) {  // not the same?
        p = nullptr;  // value is a userdata with wrong metatable
      }
      lua_pop(state, 2);  // remove both metatables
    }
    return reinterpret_cast<T*>(p);
  }
}  // namespace

void pushResourceType(lua_State* state, uint64_t rt) {
  auto& p = newUserData<uint64_t>(state);
  p = rt;
  luaL_newmetatable(state, resourceTypeKey);
  lua_setmetatable(state, -2);
}

kj::Maybe<uint64_t> getResourceType(lua_State* state, int index) {
  return testUserData<uint64_t>(state, index, resourceTypeKey);
}

namespace {
  struct IdHolder {
    kj::Own<const Id> id;
  };

  int destroyIdHolder(lua_State* state) {
    auto& holder = KJ_ASSERT_NONNULL(testUserData<IdHolder>(state, 1, idKey));
    holder.id = nullptr;
    return 0;
  }
}  // namespace

void pushId(lua_State* state, kj::Own<const Id> id) {
  auto& p = newUserData<IdHolder>(state);
  p.id = kj::mv(id);
  if (luaL_newmetatable(state, idKey)) {
    lua_pushcfunction(state, destroyIdHolder);
    lua_setfield(state, -2, "__gc");
  }
  lua_setmetatable(state, -2);
}

kj::Maybe<const Id&> getId(lua_State* state, int index) {
  return testUserData<IdHolder>(state, index, idKey).map([] (IdHolder& holder) -> const Id& {
    return *holder.id;
  });
}

}  // namespace luacat
}  // namespace mcm
