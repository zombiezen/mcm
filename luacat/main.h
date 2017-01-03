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

#ifndef MCM_LUACAT_MAIN_H_
#define MCM_LUACAT_MAIN_H_

#include "kj/common.h"
#include "kj/io.h"
#include "kj/main.h"
#include "kj/string.h"
#include "capnp/message.h"

extern "C" {
#include "lua.h"
}

namespace mcm {

namespace luacat {

class Main {
public:
  Main(kj::ProcessContext& context, kj::OutputStream& outStream, kj::OutputStream& logStream);
  KJ_DISALLOW_COPY(Main);

  kj::MainBuilder::Validity processFile(kj::StringPtr src);

  void process(capnp::MessageBuilder& out, kj::StringPtr chunkName, kj::InputStream& stream);
  // Run the Lua file from the given stream.

  kj::MainFunc getMain();

private:
  kj::ProcessContext& context;
  kj::OutputStream& outStream;
  kj::OutputStream& logStream;
};

class OwnState {
  // A transferrable title to a lua_State.
  // Similar to kj::Own, but kj::Own requires a complete type for its disposers.
  // TODO(someday): use kj::Own instead

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

OwnState newLuaState();
// Create a new Lua interpreter.

}  // namespace luacat
}  // namespace mcm

#endif  // MCM_LUACAT_MAIN_H_
