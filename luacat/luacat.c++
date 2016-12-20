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

#include <unistd.h>

#include "kj/main.h"
#include "kj/io.h"
#include "kj/string.h"
#include "capnp/message.h"
#include "capnp/serialize.h"
#include "luacat/interp.h"

namespace mcm {

namespace luacat {

class Main {
public:
  Main(kj::ProcessContext& context): context(context) {}

  kj::MainBuilder::Validity processFile(kj::StringPtr src) {
    if (src.size() == 0) {
      return kj::str("empty source");
    }
    kj::FdOutputStream out(STDOUT_FILENO);
    kj::FdOutputStream err(STDERR_FILENO);
    Lua l(err);
    l.exec(src);
    capnp::MallocMessageBuilder message;
    l.finish(message);
    capnp::writeMessage(out, message);

    return true;
  }

  kj::MainFunc getMain() {
    return kj::MainBuilder(context, "mcm-luacat", "Interprets Lua source and generates an mcm catalog.")
        .expectArg("FILE", KJ_BIND_METHOD(*this, processFile))
        .build();
  }

private:
  kj::ProcessContext& context;
};

}  // namespace luacat
}  // namespace mcm

KJ_MAIN(mcm::luacat::Main)
