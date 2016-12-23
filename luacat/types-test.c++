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

#include <iostream>
#include "gtest/gtest.h"
#include "kj/memory.h"
#include "kj/string.h"
#include "capnp/dynamic.h"
#include "lua.hpp"

#include "luacat/testsuite.capnp.h"

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

namespace {
  class OwnState {
    // A transferrable title to a lua_State. 
    // Similar to kj::Own, but kj::Own requires a complete type for its disposers.
    // TODO(someday): use kj::Own instead
    // TODO(someday): rewrite app code to use this

  public:
    KJ_DISALLOW_COPY(OwnState);
    inline OwnState(): ptr(nullptr) {}
    inline OwnState(OwnState&& other) noexcept
        : ptr(other.ptr) { other.ptr = nullptr; }
    explicit inline OwnState(lua_State* ptr) noexcept: ptr(ptr) {}

    ~OwnState() noexcept {
      if (ptr == nullptr) {
        return;
      }
      lua_close(ptr);
      ptr = nullptr;
    }

    inline OwnState& operator=(OwnState&& other) {
      // Move-assignment operator.

      lua_State* ptrCopy = ptr;
      ptr = other.ptr;
      other.ptr = nullptr;
      if (ptrCopy != nullptr) {
        lua_close(ptrCopy);
      }
      return *this;
    }

    inline OwnState& operator=(decltype(nullptr)) {
      lua_close(ptr);
      ptr = nullptr;
      return *this;
    }

    inline lua_State* get() { return ptr; }
    inline operator lua_State*() { return ptr; }

  private:
    lua_State* ptr;
  };
}  // namespace

OwnState newLuaState() {
  return OwnState(luaL_newstate());
}

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
  OwnState state = newLuaState();
  ASSERT_NO_FATAL_FAILURE(evalString(state, "{void = true}"));
  capnp::MallocMessageBuilder message;
  auto root = message.initRoot<mcm::luacat::GenericValue>();

  copyStruct(state, root);

  ASSERT_EQ(mcm::luacat::GenericValue::VOID, root.which());
}

TEST(CopyStructTest, BoolFieldTrue) {
  OwnState state = newLuaState();
  ASSERT_NO_FATAL_FAILURE(evalString(state, "{bool = true}"));
  capnp::MallocMessageBuilder message;
  auto root = message.initRoot<mcm::luacat::GenericValue>();

  copyStruct(state, root);

  ASSERT_EQ(mcm::luacat::GenericValue::BOOL, root.which());
  ASSERT_EQ(true, root.getBool());
}

TEST(CopyStructTest, BoolFieldFalse) {
  OwnState state = newLuaState();
  ASSERT_NO_FATAL_FAILURE(evalString(state, "{bool = false}"));
  capnp::MallocMessageBuilder message;
  auto root = message.initRoot<mcm::luacat::GenericValue>();

  copyStruct(state, root);

  ASSERT_EQ(mcm::luacat::GenericValue::BOOL, root.which());
  ASSERT_EQ(false, root.getBool());
}

TEST(CopyStructTest, EnumField) {
  OwnState state = newLuaState();
  ASSERT_NO_FATAL_FAILURE(evalString(state, "{enum = \"that\"}"));
  capnp::MallocMessageBuilder message;
  auto root = message.initRoot<mcm::luacat::GenericValue>();

  copyStruct(state, root);

  ASSERT_EQ(mcm::luacat::GenericValue::ENUM, root.which());
  ASSERT_EQ(mcm::luacat::Subject::THAT, root.getEnum());
}

TEST(CopyStructTest, Int64Field) {
  OwnState state = newLuaState();
  ASSERT_NO_FATAL_FAILURE(evalString(state, "{int64 = -0x7fffffffffffffff}"));
  capnp::MallocMessageBuilder message;
  auto root = message.initRoot<mcm::luacat::GenericValue>();

  copyStruct(state, root);

  ASSERT_EQ(mcm::luacat::GenericValue::INT64, root.which());
  ASSERT_EQ(-0x7fffffffffffffff, root.getInt64());
}

TEST(CopyStructTest, UInt64Field) {
  OwnState state = newLuaState();
  ASSERT_NO_FATAL_FAILURE(evalString(state, "{uint64 = 0x8000000000000000}"));
  capnp::MallocMessageBuilder message;
  auto root = message.initRoot<mcm::luacat::GenericValue>();

  copyStruct(state, root);

  ASSERT_EQ(mcm::luacat::GenericValue::UINT64, root.which());
  ASSERT_EQ(0x8000000000000000, root.getUint64());
}

TEST(CopyStructTest, TextField) {
  OwnState state = newLuaState();
  ASSERT_NO_FATAL_FAILURE(evalString(state, "{text = \"Hello, World!\"}"));
  capnp::MallocMessageBuilder message;
  auto root = message.initRoot<mcm::luacat::GenericValue>();

  copyStruct(state, root);

  ASSERT_EQ(mcm::luacat::GenericValue::TEXT, root.which());
  ASSERT_STREQ("Hello, World!", root.getText().cStr());
}

TEST(CopyStructTest, DataField) {
  OwnState state = newLuaState();
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
  OwnState state = newLuaState();
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
  OwnState state = newLuaState();
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
  OwnState state = newLuaState();
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
  OwnState state = newLuaState();
  ASSERT_NO_FATAL_FAILURE(evalString(state, "{enumList = {\"that\", \"this\"}}"));
  capnp::MallocMessageBuilder message;
  auto root = message.initRoot<mcm::luacat::GenericValue>();

  copyStruct(state, root);

  ASSERT_EQ(mcm::luacat::GenericValue::ENUM_LIST, root.which());
  ASSERT_EQ(2, root.getEnumList().size());
  EXPECT_EQ(mcm::luacat::Subject::THAT, root.getEnumList()[0]);
  EXPECT_EQ(mcm::luacat::Subject::THIS, root.getEnumList()[1]);
}
