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
  struct Reader {
    static const int bufSize = 4096;

    kj::InputStream& stream;
    kj::byte buf[bufSize];

    explicit Reader(kj::InputStream& s) : stream(s) {}
  };

  const char* readStream(lua_State* state, void* data, size_t* size) {
    auto& reader = *reinterpret_cast<Reader*>(data);
    *size = reader.stream.tryRead(reader.buf, 1, Reader::bufSize);
    return reinterpret_cast<char*>(reader.buf);
  }
}  // namespace

const kj::ArrayPtr<const kj::byte> luaBytePtr(lua_State* state, int index) {
  size_t len = 0;
  auto s = reinterpret_cast<const kj::byte*>(lua_tolstring(state, index, &len));
  return kj::ArrayPtr<const kj::byte>(s, len);
}

const kj::StringPtr luaStringPtr(lua_State* state, int index) {
  size_t len = 0;
  const char* s = lua_tolstring(state, index, &len);
  return kj::StringPtr(s, len);
}

int luaLoad(lua_State* state, kj::StringPtr name, kj::InputStream& stream) {
  Reader reader(stream);
  return lua_load(state, readStream, &reader, name.cStr(), NULL);
}

void copyStruct(lua_State* state, capnp::DynamicStruct::Builder builder) {
  if (!lua_istable(state, -1)) {
    luaL_error(state, "copyStruct: not a table");
    return;
  }
  if (!lua_checkstack(state, 2)) {
    luaL_error(state, "copyStruct: recursion depth exceeded");
    return;
  }
  lua_pushnil(state);
  while (lua_next(state, -2)) {
    if (!lua_isstring(state, -2)) {
      luaL_error(state, "copyStruct: non-string key in table");
      return;
    }
    auto key = luaStringPtr(state, -2);

    KJ_IF_MAYBE(field, builder.getSchema().findFieldByName(key)) {
      switch (field->getType().which()) {
      case capnp::schema::Type::TEXT:
        if (lua_isstring(state, -1)) {
          capnp::DynamicValue::Reader val(lua_tostring(state, -1));
          builder.set(*field, val);
        } else {
          luaL_error(state, "copyStruct: non-string value for field %s", key);
          return;
        }
        break;
      case capnp::schema::Type::DATA:
        if (lua_isstring(state, -1)) {
          capnp::DynamicValue::Reader val(luaBytePtr(state, -1));
          builder.set(*field, val);
        } else {
          luaL_error(state, "copyStruct: non-data value for field %s", key);
          return;
        }
        break;
      case capnp::schema::Type::STRUCT:
        if (lua_istable(state, -1)) {
          auto sub = builder.init(*field).as<capnp::DynamicStruct>();
          copyStruct(state, sub);
        } else {
          luaL_error(state, "copyStruct: non-struct value for field %s", key);
          return;
        }
        break;
      case capnp::schema::Type::LIST:
        if (lua_istable(state, -1)) {
          lua_len(state, -1);
          lua_Integer n = lua_tointeger(state, -1);
          lua_pop(state, 1);
          auto sub = builder.init(*field, n).as<capnp::DynamicList>();
          if (n > 0) {
            copyList(state, sub);
          }
        } else {
          luaL_error(state, "copyStruct: non-list value for field %s", key);
          return;
        }
        break;
      // TODO(soon): all the other types
      default:
        luaL_error(state, "copyStruct: can't map field %s type %d to Lua", key, field->getType().which());
        return;
      }
    } else {
      luaL_error(state, "copyStruct: unknown field '%s' in table", key);
      return;
    }

    lua_pop(state, 1);  // pop value, now key is on top.
  }
}

void copyList(lua_State* state, capnp::DynamicList::Builder builder) {
  if (!lua_checkstack(state, 2)) {
    luaL_error(state, "copyList: recursion depth exceeded");
    return;
  }
  switch (builder.getSchema().whichElementType()) {
  case capnp::schema::Type::UINT64:
    for (lua_Integer i = 0; i < builder.size(); i++) {
      if (lua_geti(state, -1, i + 1) != LUA_TNUMBER) {
        luaL_error(state, "copyList: found non-number in List(UInt64)");
        return;
      }
      capnp::DynamicValue::Reader val(static_cast<uint64_t>(lua_tointeger(state, -1)));
      builder.set(i, val);
      lua_pop(state, 1);
    }
    break;
  case capnp::schema::Type::TEXT:
    for (lua_Integer i = 0; i < builder.size(); i++) {
      if (lua_geti(state, -1, i + 1) != LUA_TSTRING) {
        luaL_error(state, "copyList: found non-string in List(Text)");
        return;
      }
      capnp::DynamicValue::Reader val(lua_tostring(state, -1));
      builder.set(i, val);
      lua_pop(state, 1);
    }
    break;
  case capnp::schema::Type::STRUCT:
    for (lua_Integer i = 0; i < builder.size(); i++) {
      if (lua_geti(state, -1, i) != LUA_TTABLE) {
        luaL_error(state, "copyList: found non-table in List(Text)");
        return;
      }
      copyStruct(state, builder[i].as<capnp::DynamicStruct>());
      lua_pop(state, 1);
    }
    break;
  default:
    luaL_error(state, "copyList: can't map type %d to Lua", builder.getSchema().whichElementType());
    return;
  }
}

}  // namespace luacat
}  // namespace mcm
