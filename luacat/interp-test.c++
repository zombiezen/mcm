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

#include "luacat/interp.h"

#include <iostream>
#include "gtest/gtest.h"
#include "kj/io.h"
#include "kj/string.h"

#include "luacat/testsuite.capnp.h"

namespace kj {
  inline void PrintTo(const kj::String& s, ::std::ostream* os) {
    os->write(s.begin(), s.size());
  }

  inline void PrintTo(capnp::Text::Reader s, ::std::ostream* os) {
    os->write(s.begin(), s.size());
  }

  inline void PrintTo(kj::StringPtr s, ::std::ostream* os) {
    os->write(s.begin(), s.size());
  }
}

const int logBufMax = 4096;

TEST(LuaTest, NoExecHasEmptyCatalog) {
  auto logBuf = kj::heapArray<kj::byte>(logBufMax);
  kj::ArrayOutputStream logBufStream(logBuf);
  mcm::luacat::Lua l(logBufStream);
  capnp::MallocMessageBuilder message;
  l.finish(message);

  auto catalog = message.getRoot<mcm::Catalog>().asReader();
  ASSERT_EQ(0, catalog.getResources().size());
}

TEST(LuaTest, TestSuite) {
  for (auto testCase : mcm::luacat::TEST_SUITE->getTests()) {
    SCOPED_TRACE(testCase.getName().cStr());
    auto logBuf = kj::heapArray<kj::byte>(logBufMax);
    kj::ArrayOutputStream logBufStream(logBuf);
    mcm::luacat::Lua l(logBufStream);
    // TODO(soon): catch exceptions
    kj::ArrayInputStream scriptStream(testCase.getScript().asBytes());
    l.exec("=(load)", scriptStream);
    capnp::MallocMessageBuilder message;
    l.finish(message);

    // TODO(soon): check catalogs for equality, if need be
    auto outArray = logBufStream.getArray();
    auto outString = kj::heapString(reinterpret_cast<char*>(outArray.begin()), outArray.size());
    EXPECT_EQ(testCase.getExpected().getOutput(), outString);
  }
}
