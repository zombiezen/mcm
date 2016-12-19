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

namespace mcm {

namespace luacat {

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

}  // namespace luacat
}  // namespace mcm
