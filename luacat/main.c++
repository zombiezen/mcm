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

#include "luacat/main.h"

#include <unistd.h>
#include <fcntl.h>
#include "kj/debug.h"
#include "kj/exception.h"
#include "capnp/serialize.h"

extern "C" {
#include "lauxlib.h"
#include "lualib.h"
}

#include "luacat/convert.h"
#include "luacat/lib.h"
#include "luacat/path.h"

namespace mcm {

namespace luacat {

namespace {
  const luaL_Reg loadedlibs[] = {
    {"_G", luaopen_base},
    {LUA_LOADLIBNAME, luaopen_package},
    {LUA_COLIBNAME, luaopen_coroutine},
    {LUA_TABLIBNAME, luaopen_table},
    {LUA_STRLIBNAME, luaopen_string},
    {LUA_MATHLIBNAME, luaopen_math},
    {LUA_UTF8LIBNAME, luaopen_utf8},
    {NULL, NULL}
  };

  bool isValidLuaInclude(const kj::ArrayPtr<const char> path) {
    for (char c: path) {
      if (c == '?') {
        return true;
      }
    }
    return false;
  }

  int printfunc(lua_State *state) {
    // Customized implementation of print().
    // We could customize this in vendored copy, but this keeps the
    // application/vendored code separation clean.

    auto& stream = *reinterpret_cast<kj::OutputStream*>(lua_touserdata(state, lua_upvalueindex(1)));
    int n = lua_gettop(state);  // number of arguments
    int i;
    lua_getglobal(state, "tostring");
    for (i = 1; i <= n; i++) {
      const char *s;
      size_t l;
      lua_pushvalue(state, -1);  // function to be called
      lua_pushvalue(state, i);   // value to print
      lua_call(state, 1, 1);
      s = lua_tolstring(state, -1, &l);  // get result
      if (s == NULL) {
        return luaL_error(state, "'tostring' must return a string to 'print'");
      }
      if (i > 1) {
        stream.write("\t", 1);
      }
      stream.write(s, l);
      lua_pop(state, 1);  // pop result
    }
    stream.write("\n", 1);
    return 0;
  }
}  // namespace

Main::Main(kj::ProcessContext& context, kj::OutputStream& outStream, kj::OutputStream& logStream):
    context(context), outStream(&outStream), logStream(logStream) {
}

void Main::setFallbackIncludePath(kj::StringPtr include) {
  if (include.size() == 0) {
    fallbackInclude = kj::heapString(include);
    return;
  }
  auto parts = splitStr(include, ';');
  kj::StringTree cleaned;
  for (auto part: parts) {
    if (!isValidLuaInclude(part)) {
      continue;
    }
    if (cleaned.size() > 0) {
      cleaned = kj::strTree(kj::mv(cleaned), ";");
    }
    cleaned = kj::strTree(kj::mv(cleaned), part);
  }
  fallbackInclude = cleaned.flatten();
}

kj::MainBuilder::Validity Main::addIncludePath(kj::StringPtr include) {
  auto parts = splitStr(include, ';');
  for (auto part: parts) {
    if (!isValidLuaInclude(part)) {
      return kj::str("path '", part, "' does not include a '?' wildcard");
    }
  }
  if (includes.size() == 0) {
    includes = kj::strTree(include);
    return true;
  }
  includes = kj::strTree(kj::mv(includes), ";", include);
  return true;
}

kj::MainBuilder::Validity Main::setOutputPath(kj::StringPtr outPath) {
  int fd;
  KJ_SYSCALL(fd = open(outPath.cStr(), O_WRONLY | O_CREAT | O_TRUNC, 0666), outPath);
  kj::AutoCloseFd autoclose(fd);
  ownOutStream = kj::heap<kj::FdOutputStream>(kj::mv(autoclose));
  outStream = ownOutStream;
  return true;
}

kj::MainBuilder::Validity Main::processFile(kj::StringPtr src) {
  if (src.size() == 0) {
    return kj::str("empty source");
  }
  auto maybeFdStream = kj::dynamicDowncastIfAvailable<kj::FdOutputStream, kj::OutputStream>(*outStream);
  KJ_IF_MAYBE(f, maybeFdStream) {
    if (isatty(f->getFd())) {
      context.exitError("mcm-luacat: output file is a tty\n\nWriting a binary catalog will likely mess up your terminal. Either\nredirect stdout or use -o.");
    }
  }
  auto chunkName = kj::str("@", src);
  auto maybeExc = kj::runCatchingExceptions([&]() {
    int fd;
    KJ_SYSCALL(fd = open(src.cStr(), O_RDONLY, 0), src);
    kj::AutoCloseFd afd(fd);
    kj::FdInputStream stream(kj::mv(afd));
    capnp::MallocMessageBuilder message;
    process(message, chunkName, stream);
    capnp::writeMessage(*outStream, message);
  });
  KJ_IF_MAYBE(e, maybeExc) {
    context.error(e->getDescription());
  }

  return true;
}

void Main::process(capnp::MessageBuilder& message, kj::StringPtr chunkName, kj::InputStream& stream) {
  auto state = newLuaState();

  // Load libraries
  const luaL_Reg *reg;
  for (reg = loadedlibs; reg->func; reg++) {
    luaL_requiref(state, reg->name, reg->func, 1);
    lua_pop(state, 1);  // remove lib
  }
  LibState libState;
  openlib(state, libState);  // push mcm module
  lua_setglobal(state, "mcm");  // _G.mcm = module

  // Override print function.
  lua_getglobal(state, "_G");
  lua_pushlightuserdata(state, &logStream);
  lua_pushcclosure(state, printfunc, 1);
  lua_setfield(state, -2, "print");
  lua_pop(state, 1);

  // Set package.path
  {
    auto inc = buildIncludePath(chunkName);
    lua_getglobal(state, "package");
    pushLua(state, inc);
    lua_setfield(state, -2, "path");
    lua_pop(state, 1);
  }

  // Run script
  if (luaLoad(state, chunkName, stream) || lua_pcall(state, 0, 0, 0)) {
    auto errMsg = kj::heapString(luaStringPtr(state, -1));
    lua_pop(state, 1);
    throw kj::Exception(kj::Exception::Type::FAILED, __FILE__, __LINE__, kj::mv(errMsg));
  }

  // Create catalog
  auto catalog = message.initRoot<Catalog>();
  auto resources = libState.getResources();
  auto rlist = catalog.initResources(resources.size());
  // TODO(soon): sort
  for (size_t i = 0; i < resources.size(); i++) {
    rlist.setWithCaveats(i, resources[i].get());
  }
}

kj::String Main::buildIncludePath(kj::StringPtr chunkName) {
  kj::StringTree tree;
  if (chunkName.startsWith("@")) {
    // Actual file name; add containing directory.
    auto srcDir = dirName(chunkName.slice(1));
    tree = kj::strTree(joinPath(srcDir, "?.lua"), ";", joinPath(srcDir, "?", "init.lua"));
  }
  if (includes.size() > 0) {
    if (tree.size() > 0) {
      tree = kj::strTree(kj::mv(tree), ";");
    }
    tree = kj::strTree(kj::mv(tree), includes.flatten());
  }
  if (fallbackInclude.size() > 0) {
    if (tree.size() > 0) {
      tree = kj::strTree(kj::mv(tree), ";");
    }
    tree = kj::strTree(kj::mv(tree), fallbackInclude);
  }
  return tree.flatten();
}

kj::MainFunc Main::getMain() {
  return kj::MainBuilder(context, "mcm-luacat", "Interprets Lua source and generates an mcm catalog.")
      .addOptionWithArg({'I'}, KJ_BIND_METHOD(*this, addIncludePath),
          "<templates>", "Add a package path template in package.searchpath format.")
      .addOptionWithArg({'o'}, KJ_BIND_METHOD(*this, setOutputPath),
          "FILE", "Write output to FILE instead of stdout.")
      .expectArg("FILE", KJ_BIND_METHOD(*this, processFile))
      .build();
}

OwnState newLuaState() {
  // TODO(someday): use lua_newstate and set atpanic
  lua_State* state = luaL_newstate();
  KJ_ASSERT_NONNULL(state);
  return OwnState(state);
}

}  // namespace luacat
}  // namespace mcm
