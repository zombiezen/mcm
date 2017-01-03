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

#include "luacat/main.h"

#include <iostream>
#include "gtest/gtest.h"
#include "capnp/any.h"
#include "kj/debug.h"
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

namespace capnp {
  inline void PrintTo(capnp::Equality eq, ::std::ostream* os) {
    auto s = kj::str(eq);
    kj::PrintTo(s, os);
  }
}

namespace {
  struct NullInputStream : public kj::InputStream {
    size_t tryRead(void* buffer, size_t minBytes, size_t maxBytes) override {
      return 0;
    }
  };

  struct DiscardOutputStream : public kj::OutputStream {
    void write(const void* buffer, size_t size) override {
    }
  };

  struct FakeProcessContext : public kj::ProcessContext {
    kj::StringPtr getProgramName() override {
      return nullptr;
    }

    void exit() override {
      KJ_FAIL_ASSERT("exit");
    }

    void warning(kj::StringPtr message) override {}

    void error(kj::StringPtr message) override {}

    void exitError(kj::StringPtr message) override {
      error(message);
      exit();
    }

    void exitInfo(kj::StringPtr message) override {
      exit();
    }

    void increaseLoggingVerbosity() override {}
  };
}  // namespace

const int logBufMax = 4096;

TEST(MainTest, TestSuite) {
  for (auto testCase : mcm::luacat::TEST_SUITE->getTests()) {
    SCOPED_TRACE(testCase.getName().cStr());
    DiscardOutputStream discardStdout;
    auto logBuf = kj::heapArray<kj::byte>(logBufMax);
    kj::ArrayOutputStream logBufStream(logBuf);
    FakeProcessContext ctx;
    mcm::luacat::Main main(ctx, discardStdout, logBufStream);
    // TODO(soon): catch exceptions
    kj::ArrayInputStream scriptStream(testCase.getScript().asBytes());
    capnp::MallocMessageBuilder message;
    main.process(message, "=(load)", scriptStream);

    if (testCase.getExpected().hasCatalog()) {
      auto catalog = message.getRoot<mcm::Catalog>().asReader();
      auto catalogStr = kj::str(catalog);
      capnp::AnyStruct::Reader catalogAny(catalog);
      auto wantCatalog = testCase.getExpected().getCatalog();
      auto wantCatalogStr = kj::str(wantCatalog);
      capnp::AnyStruct::Reader wantCatalogAny(wantCatalog);
      EXPECT_EQ(capnp::Equality::EQUAL, catalogAny.equals(wantCatalogAny)) << "got:  " << catalogStr.cStr() << "\nwant: " << wantCatalogStr.cStr();
    }
    auto outArray = logBufStream.getArray();
    auto outString = kj::heapString(reinterpret_cast<char*>(outArray.begin()), outArray.size());
    EXPECT_EQ(testCase.getExpected().getOutput(), outString);
  }
}
