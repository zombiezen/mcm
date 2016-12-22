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

#include "kj/debug.h"
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
  KJ_ASSERT(lua_checkstack(state, 2), "recursion depth exceeded");
  auto structName = builder.getSchema().getShortDisplayName();
  KJ_CONTEXT(structName);
  KJ_REQUIRE(lua_istable(state, -1), "value must be a table");
  lua_pushnil(state);
  while (lua_next(state, -2)) {
    KJ_REQUIRE(lua_isstring(state, -2), "non-string key in table");
    auto key = luaStringPtr(state, -2);
    KJ_CONTEXT(key);

    auto field = KJ_REQUIRE_NONNULL(builder.getSchema().findFieldByName(key), "could not find field");
    switch (field.getType().which()) {
    case capnp::schema::Type::VOID:
      {
        capnp::DynamicValue::Reader val(capnp::VOID);
        builder.set(field, val);
      }
      break;
    case capnp::schema::Type::BOOL:
      {
        KJ_REQUIRE(lua_isboolean(state, -1), "non-boolean value");
        capnp::DynamicValue::Reader val(static_cast<bool>(lua_toboolean(state, -1)));
        builder.set(field, val);
      }
      break;
    case capnp::schema::Type::INT8:
    case capnp::schema::Type::INT16:
    case capnp::schema::Type::INT32:
    case capnp::schema::Type::INT64:
      {
        KJ_REQUIRE(lua_isnumber(state, -1), "non-number value");
        int isint = 0;
        capnp::DynamicValue::Reader val(static_cast<int64_t>(lua_tointegerx(state, -1, &isint)));
        KJ_REQUIRE(isint, "non-integer value");
        builder.set(field, val);
      }
      break;
    case capnp::schema::Type::UINT8:
    case capnp::schema::Type::UINT16:
    case capnp::schema::Type::UINT32:
    case capnp::schema::Type::UINT64:
      {
        KJ_REQUIRE(lua_isnumber(state, -1), "non-number value");
        int isint = 0;
        capnp::DynamicValue::Reader val(static_cast<uint64_t>(lua_tointegerx(state, -1, &isint)));
        KJ_REQUIRE(isint, "non-integer value");
        builder.set(field, val);
      }
      break;
    case capnp::schema::Type::FLOAT32:
    case capnp::schema::Type::FLOAT64:
      {
        KJ_REQUIRE(lua_isnumber(state, -1), "non-number value");
        capnp::DynamicValue::Reader val(lua_tonumber(state, -1));
        builder.set(field, val);
      }
      break;
    case capnp::schema::Type::TEXT:
      {
        KJ_REQUIRE(lua_isstring(state, -1), "non-string value");
        capnp::DynamicValue::Reader val(luaStringPtr(state, -1));
        builder.set(field, val);
      }
      break;
    case capnp::schema::Type::DATA:
      {
        KJ_REQUIRE(lua_isstring(state, -1), "non-string value");
        capnp::DynamicValue::Reader val(luaBytePtr(state, -1));
        builder.set(field, val);
      }
      break;
    case capnp::schema::Type::LIST:
      {
        KJ_REQUIRE(lua_istable(state, -1), "non-table value");
        lua_len(state, -1);
        lua_Integer n = lua_tointeger(state, -1);
        lua_pop(state, 1);
        auto sub = builder.init(field, n).as<capnp::DynamicList>();
        if (n > 0) {
          copyList(state, sub);
        }
      }
      break;
    case capnp::schema::Type::ENUM:
      {
        KJ_REQUIRE(lua_isstring(state, -1), "non-string value");
        auto sval = luaStringPtr(state, -1);
        auto schema = field.getType().asEnum();
        auto e = KJ_REQUIRE_NONNULL(schema.findEnumerantByName(sval), "could not find enum value", sval);
        capnp::DynamicValue::Reader val(e);
        builder.set(field, val);
      }
      break;
    case capnp::schema::Type::STRUCT:
      {
        KJ_REQUIRE(lua_istable(state, -1), "non-table value");
        auto sub = builder.init(field).as<capnp::DynamicStruct>();
        copyStruct(state, sub);
      }
      break;
    default:
      KJ_FAIL_REQUIRE("can't map field type to Lua", field.getType().which());
    }

    lua_pop(state, 1);  // pop value, now key is on top.
  }
}

void copyList(lua_State* state, capnp::DynamicList::Builder builder) {
  KJ_ASSERT(lua_checkstack(state, 2), "recursion depth exceeded");
  switch (builder.getSchema().whichElementType()) {
  case capnp::schema::Type::VOID:
    // Do nothing.
    break;
  case capnp::schema::Type::BOOL:
    for (lua_Integer i = 0; i < builder.size(); i++) {
      KJ_CONTEXT("List(Bool)", i);
      int ty = lua_geti(state, -1, i + 1);
      KJ_REQUIRE(ty == LUA_TBOOLEAN, "non-boolean element");
      capnp::DynamicValue::Reader val(static_cast<bool>(lua_toboolean(state, -1)));
      builder.set(i, val);
      lua_pop(state, 1);
    }
    break;
  // TODO(#6): rest of list types
  case capnp::schema::Type::UINT64:
    for (lua_Integer i = 0; i < builder.size(); i++) {
      KJ_CONTEXT("List(UInt64)", i);
      int ty = lua_geti(state, -1, i + 1);
      KJ_REQUIRE(ty == LUA_TNUMBER, "non-number element");
      capnp::DynamicValue::Reader val(static_cast<uint64_t>(lua_tointeger(state, -1)));
      builder.set(i, val);
      lua_pop(state, 1);
    }
    break;
  case capnp::schema::Type::TEXT:
    for (lua_Integer i = 0; i < builder.size(); i++) {
      KJ_CONTEXT("List(Text)", i);
      int ty = lua_geti(state, -1, i + 1);
      KJ_REQUIRE(ty == LUA_TSTRING, "non-string element");
      capnp::DynamicValue::Reader val(luaStringPtr(state, -1));
      builder.set(i, val);
      lua_pop(state, 1);
    }
    break;
  case capnp::schema::Type::STRUCT:
    {
      auto structName = builder.getSchema().getStructElementType().getShortDisplayName();
      for (lua_Integer i = 0; i < builder.size(); i++) {
        KJ_CONTEXT("List(Struct)", i, structName);
        int ty = lua_geti(state, -1, i + 1);
        KJ_REQUIRE(ty == LUA_TTABLE, "non-table element");
        copyStruct(state, builder[i].as<capnp::DynamicStruct>());
        lua_pop(state, 1);
      }
    }
    break;
  default:
    KJ_FAIL_REQUIRE("can't map type to Lua", builder.getSchema().whichElementType());
  }
}

}  // namespace luacat
}  // namespace mcm
