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

#include "luacat/convert.h"

#include <iostream>
#include "gtest/gtest.h"
#include "kj/memory.h"
#include "kj/string.h"
#include "capnp/dynamic.h"
#include "lua.hpp"

#include "luacat/main.h"  // for OwnState
#include "luacat/testsuite.capnp.h"
#include "luacat/types.h"

namespace kj {
  void PrintTo(kj::ArrayPtr<const kj::byte> s, ::std::ostream* os) {
    if (s.size() == 0) {
      return;
    }
    auto builder = kj::heapArrayBuilder<char>(s.size()*3 - 1);
    for (size_t i = 0; i < s.size(); i++) {
      if (i > 0) {
        builder.add(' ');
      }
      builder.addAll(kj::hex(s[i]));
    }
    os->write(builder.begin(), builder.size());
  }
}  // namespace kj

void evalString(lua_State* state, kj::StringPtr s) {
  SCOPED_TRACE("evalString");
  auto actual = kj::str("return ", s, "\n");
  int result = luaL_loadstring(state, actual.cStr());
  ASSERT_EQ(result, LUA_OK) << "failed to compile: " << s.cStr();
  result = lua_pcall(state, 0, 1, 0);
  if (result != LUA_OK) {
    const char* msg = lua_tostring(state, -1);
    FAIL() << "Lua error: " << msg;
  }
}

TEST(CopyStructTest, VoidField) {
  auto state = mcm::luacat::newLuaState();
  ASSERT_NO_FATAL_FAILURE(evalString(state, "{void = true}"));
  capnp::MallocMessageBuilder message;
  auto root = message.initRoot<mcm::luacat::GenericValue>();

  copyStruct(state, root);

  ASSERT_EQ(mcm::luacat::GenericValue::VOID, root.which());
}

TEST(CopyStructTest, BoolFieldTrue) {
  auto state = mcm::luacat::newLuaState();
  ASSERT_NO_FATAL_FAILURE(evalString(state, "{bool = true}"));
  capnp::MallocMessageBuilder message;
  auto root = message.initRoot<mcm::luacat::GenericValue>();

  copyStruct(state, root);

  ASSERT_EQ(mcm::luacat::GenericValue::BOOL, root.which());
  ASSERT_EQ(true, root.getBool());
}

TEST(CopyStructTest, BoolFieldFalse) {
  auto state = mcm::luacat::newLuaState();
  ASSERT_NO_FATAL_FAILURE(evalString(state, "{bool = false}"));
  capnp::MallocMessageBuilder message;
  auto root = message.initRoot<mcm::luacat::GenericValue>();

  copyStruct(state, root);

  ASSERT_EQ(mcm::luacat::GenericValue::BOOL, root.which());
  ASSERT_EQ(false, root.getBool());
}

TEST(CopyStructTest, EnumField) {
  auto state = mcm::luacat::newLuaState();
  ASSERT_NO_FATAL_FAILURE(evalString(state, "{enum = \"that\"}"));
  capnp::MallocMessageBuilder message;
  auto root = message.initRoot<mcm::luacat::GenericValue>();

  copyStruct(state, root);

  ASSERT_EQ(mcm::luacat::GenericValue::ENUM, root.which());
  ASSERT_EQ(mcm::luacat::Subject::THAT, root.getEnum());
}

TEST(CopyStructTest, Int64Field) {
  auto state = mcm::luacat::newLuaState();
  ASSERT_NO_FATAL_FAILURE(evalString(state, "{int64 = -0x7fffffffffffffff}"));
  capnp::MallocMessageBuilder message;
  auto root = message.initRoot<mcm::luacat::GenericValue>();

  copyStruct(state, root);

  ASSERT_EQ(mcm::luacat::GenericValue::INT64, root.which());
  ASSERT_EQ(-0x7fffffffffffffff, root.getInt64());
}

TEST(CopyStructTest, UInt64Field) {
  auto state = mcm::luacat::newLuaState();
  ASSERT_NO_FATAL_FAILURE(evalString(state, "{uint64 = 0x8000000000000000}"));
  capnp::MallocMessageBuilder message;
  auto root = message.initRoot<mcm::luacat::GenericValue>();

  copyStruct(state, root);

  ASSERT_EQ(mcm::luacat::GenericValue::UINT64, root.which());
  ASSERT_EQ(0x8000000000000000, root.getUint64());
}

TEST(CopyStructTest, UInt64FieldWithId) {
  auto state = mcm::luacat::newLuaState();
  lua_createtable(state, 0, 1);
  mcm::luacat::pushId(state, kj::heap<const mcm::luacat::Id>(42, nullptr));
  lua_setfield(state, -2, "uint64");
  capnp::MallocMessageBuilder message;
  auto root = message.initRoot<mcm::luacat::GenericValue>();

  copyStruct(state, root);

  ASSERT_EQ(mcm::luacat::GenericValue::UINT64, root.which());
  ASSERT_EQ(42, root.getUint64());
}

TEST(CopyStructTest, TextField) {
  auto state = mcm::luacat::newLuaState();
  ASSERT_NO_FATAL_FAILURE(evalString(state, "{text = \"Hello, World!\"}"));
  capnp::MallocMessageBuilder message;
  auto root = message.initRoot<mcm::luacat::GenericValue>();

  copyStruct(state, root);

  ASSERT_EQ(mcm::luacat::GenericValue::TEXT, root.which());
  ASSERT_STREQ("Hello, World!", root.getText().cStr());
}

TEST(CopyStructTest, DataField) {
  auto state = mcm::luacat::newLuaState();
  ASSERT_NO_FATAL_FAILURE(evalString(state, "{data = \"Hello, World!\"}"));
  capnp::MallocMessageBuilder message;
  auto root = message.initRoot<mcm::luacat::GenericValue>();

  copyStruct(state, root);

  ASSERT_EQ(mcm::luacat::GenericValue::DATA, root.which());
  ASSERT_EQ(
      kj::StringPtr("Hello, World!").asBytes(),
      kj::ArrayPtr<const kj::byte>(root.getData()));
}

TEST(CopyListTest, BoolList) {
  auto state = mcm::luacat::newLuaState();
  ASSERT_NO_FATAL_FAILURE(evalString(state, "{boolList = {true, false, true}}"));
  capnp::MallocMessageBuilder message;
  auto root = message.initRoot<mcm::luacat::GenericValue>();

  copyStruct(state, root);

  ASSERT_EQ(mcm::luacat::GenericValue::BOOL_LIST, root.which());
  ASSERT_EQ(3, root.getBoolList().size());
  EXPECT_EQ(true, root.getBoolList()[0]);
  EXPECT_EQ(false, root.getBoolList()[1]);
  EXPECT_EQ(true, root.getBoolList()[2]);
}

TEST(CopyListTest, StructList) {
  auto state = mcm::luacat::newLuaState();
  ASSERT_NO_FATAL_FAILURE(evalString(state, "{structList = {{bool = true}, {int64 = 42}}}"));
  capnp::MallocMessageBuilder message;
  auto root = message.initRoot<mcm::luacat::GenericValue>();

  copyStruct(state, root);

  ASSERT_EQ(mcm::luacat::GenericValue::STRUCT_LIST, root.which());
  ASSERT_EQ(2, root.getStructList().size());
  ASSERT_EQ(mcm::luacat::GenericValue::BOOL, root.getStructList()[0].which());
  ASSERT_EQ(true, root.getStructList()[0].getBool());
  ASSERT_EQ(mcm::luacat::GenericValue::INT64, root.getStructList()[1].which());
  ASSERT_EQ(42, root.getStructList()[1].getInt64());
}

TEST(CopyListTest, ListList) {
  auto state = mcm::luacat::newLuaState();
  ASSERT_NO_FATAL_FAILURE(evalString(state, "{listList = {{}, {-1, 42}, {314}}}"));
  capnp::MallocMessageBuilder message;
  auto root = message.initRoot<mcm::luacat::GenericValue>();

  copyStruct(state, root);

  ASSERT_EQ(mcm::luacat::GenericValue::LIST_LIST, root.which());
  ASSERT_EQ(3, root.getListList().size());
  EXPECT_EQ(0, root.getListList()[0].size());
  ASSERT_EQ(2, root.getListList()[1].size());
  EXPECT_EQ(-1, root.getListList()[1][0]);
  EXPECT_EQ(42, root.getListList()[1][1]);
  ASSERT_EQ(1, root.getListList()[2].size());
  EXPECT_EQ(314, root.getListList()[2][0]);
}

TEST(CopyListTest, EnumList) {
  auto state = mcm::luacat::newLuaState();
  ASSERT_NO_FATAL_FAILURE(evalString(state, "{enumList = {\"that\", \"this\"}}"));
  capnp::MallocMessageBuilder message;
  auto root = message.initRoot<mcm::luacat::GenericValue>();

  copyStruct(state, root);

  ASSERT_EQ(mcm::luacat::GenericValue::ENUM_LIST, root.which());
  ASSERT_EQ(2, root.getEnumList().size());
  EXPECT_EQ(mcm::luacat::Subject::THAT, root.getEnumList()[0]);
  EXPECT_EQ(mcm::luacat::Subject::THIS, root.getEnumList()[1]);
}

TEST(CopyListTest, UInt64List) {
  auto state = mcm::luacat::newLuaState();
  ASSERT_NO_FATAL_FAILURE(evalString(state, "{uint64List = {42, 0, 0xdeadbeef, 0x8000000000000000}}"));
  capnp::MallocMessageBuilder message;
  auto root = message.initRoot<mcm::luacat::GenericValue>();

  copyStruct(state, root);

  ASSERT_EQ(mcm::luacat::GenericValue::UINT64_LIST, root.which());
  auto list = root.getUint64List();
  ASSERT_EQ(4, list.size());
  EXPECT_EQ(42, list[0]);
  EXPECT_EQ(0, list[1]);
  EXPECT_EQ(0xdeadbeef, list[2]);
  EXPECT_EQ(0x8000000000000000, list[3]);
}

TEST(CopyListTest, UInt64ListWithId) {
  auto state = mcm::luacat::newLuaState();
  lua_createtable(state, 0, 1);
  lua_createtable(state, 1, 0);
  mcm::luacat::pushId(state, kj::heap<const mcm::luacat::Id>(42, nullptr));
  lua_seti(state, -2, 1);
  lua_setfield(state, -2, "uint64List");
  capnp::MallocMessageBuilder message;
  auto root = message.initRoot<mcm::luacat::GenericValue>();

  copyStruct(state, root);

  ASSERT_EQ(mcm::luacat::GenericValue::UINT64_LIST, root.which());
  auto list = root.getUint64List();
  ASSERT_EQ(1, list.size());
  EXPECT_EQ(42, list[0]);
}
