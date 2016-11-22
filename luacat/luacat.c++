#include <unistd.h>

#include "kj/array.h"
#include "kj/debug.h"
#include "kj/io.h"
#include "kj/main.h"
#include "kj/string.h"
#include "kj/vector.h"
#include "capnp/dynamic.h"
#include "capnp/message.h"
#include "capnp/orphan.h"
#include "capnp/schema.h"
#include "capnp/serialize.h"
#include "lua.hpp"
#include "openssl/sha.h"

#include "catalog.capnp.h"
#include "luacat/value.h"

namespace mcm {

namespace luacat {

class Lua {

public:
  Lua();
  KJ_DISALLOW_COPY(Lua);

  void exec(kj::StringPtr fname);
  Resource::Builder newResource();
  void finish(capnp::MessageBuilder& message);

  ~Lua();

private:
  kj::StringPtr toString(int index);

  lua_State* state;
  capnp::MallocMessageBuilder scratch;
  kj::Vector<capnp::Orphan<Resource>> resources;
};

namespace {
  const char* idHashPrefix = "mcm-luacat ID: ";
  const char* resourceTypeMetaKey = "mcm_resource";
  const uint64_t fileResId = 0x8dc4ac52b2962163;
  const uint64_t execResId = 0x984c97311006f1ca;

  const kj::ArrayPtr<const kj::byte> luaBytePtr(lua_State* state, int index) {
    size_t len;
    auto s = reinterpret_cast<const kj::byte*>(lua_tolstring(state, index, &len));
    return kj::ArrayPtr<const kj::byte>(s, len);
  }

  const kj::StringPtr luaStringPtr(lua_State* state, int index) {
    size_t len;
    const char* s = lua_tolstring(state, index, &len);
    return kj::StringPtr(s, len);
  }

  uint64_t idHash(kj::StringPtr s) {
    SHA_CTX ctx;
    SHA1_Init(&ctx);
    SHA1_Update(&ctx, idHashPrefix, strlen(idHashPrefix));
    SHA1_Update(&ctx, s.cStr(), s.size());
    uint8_t hash[SHA_DIGEST_LENGTH];
    SHA1_Final(hash, &ctx);
    return 1 | hash[0] |
        (((uint64_t)hash[1]) << 8) |
        (((uint64_t)hash[2]) << 16) |
        (((uint64_t)hash[3]) << 24) |
        (((uint64_t)hash[4]) << 32) |
        (((uint64_t)hash[5]) << 40) |
        (((uint64_t)hash[6]) << 48) |
        (((uint64_t)hash[7]) << 56);
  }

  int hashfunc(lua_State* state) {
    if (lua_gettop(state) != 1) {
      return luaL_error(state, "'mcm.hash' takes 1 argument, got %d", lua_gettop(state));
    }
    luaL_argcheck(state, lua_isstring(state, 1), 1, "must be a string");
    auto comment = luaStringPtr(state, 1);
    pushId(state, kj::heap<Id>(idHash(comment), comment));
    return 1;
  }

  void setResourceType(lua_State* state, int index, uint64_t val) {
    // Get or create metatable and leave it at top of stack.
    if (!lua_getmetatable(state, index)) {
      lua_createtable(state, 0, 1);
      lua_pushvalue(state, -1);
      lua_setmetatable(state, index);  // pops the table
    }

    // metatable[resourceTypeMetaKey] = resourceType(val)
    pushResourceType(state, val);
    lua_setfield(state, -2, resourceTypeMetaKey);

    lua_pop(state, 1);
  }

  int filefunc(lua_State* state) {
    if (lua_gettop(state) != 1) {
      return luaL_error(state, "'mcm.file' takes 1 argument, got %d", lua_gettop(state));
    }
    luaL_argcheck(state, lua_istable(state, 1), 1, "must be a table");
    setResourceType(state, 1, fileResId);
    return 1;  // Return original argument
  }

  int execfunc(lua_State* state) {
    if (lua_gettop(state) != 1) {
      return luaL_error(state, "'mcm.exec' takes 1 argument, got %d", lua_gettop(state));
    }
    luaL_argcheck(state, lua_istable(state, 1), 1, "must be a table");
    setResourceType(state, 1, execResId);
    return 1;  // Return original argument
  }

  void copyStruct(lua_State* state, capnp::DynamicStruct::Builder builder);
  void copyList(lua_State* state, capnp::DynamicList::Builder builder);

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

  int resourcefunc(lua_State* state) {
    if (lua_gettop(state) != 3) {
      return luaL_error(state, "'mcm.resource' takes 3 arguments, got %d", lua_gettop(state));
    }
    luaL_argcheck(state, lua_istable(state, 2), 2, "must be a table");
    luaL_argcheck(state, lua_istable(state, 3), 3, "must be a table");

    if (!luaL_getmetafield(state, 3, resourceTypeMetaKey)) {
      return luaL_argerror(state, 3, "expect resource table");
    }
    auto maybeTypeId = getResourceType(state, -1);
    uint64_t typeId;
    KJ_IF_MAYBE(t, maybeTypeId) {
      typeId = *t;
    } else {
      return luaL_argerror(state, 3, "expect resource table");
    }
    lua_pop(state, 1);

    Lua* l = reinterpret_cast<Lua*>(lua_touserdata(state, lua_upvalueindex(1)));
    auto res = l->newResource();
    KJ_IF_MAYBE(id, getId(state, 1)) {
      res.setId(id->getValue());
      res.setComment(id->getComment());
    } else if (lua_isstring(state, 1)) {
      auto comment = luaStringPtr(state, 1);
      res.setId(idHash(comment));
      res.setComment(comment);
    } else {
      return luaL_argerror(state, 1, "expect mcm.hash or string");
    }
    lua_len(state, 2);
    lua_Integer ndeps = lua_tointeger(state, -1);
    lua_pop(state, 1);
    if (ndeps > 0) {
      auto depList = res.initDependencies(ndeps);
      // TODO(soon): sort
      for (lua_Integer i = 1; i <= ndeps; i++) {
        lua_geti(state, 2, i);
        KJ_IF_MAYBE(id, getId(state, -1)) {
          depList.set(i-1, id->getValue());
        } else if (lua_isstring(state, -1)) {
          depList.set(i-1, idHash(luaStringPtr(state, -1)));
        } else {
          return luaL_argerror(state, 2, "expect deps to contain only mcm.hash or strings");
        }
        lua_pop(state, 1);
      }
    }

    switch (typeId) {
    case 0:
      res.setNoop();
      break;
    case fileResId:
      {
        auto f = res.initFile();
        copyStruct(state, f);
      }
      break;
    case execResId:
      {
        auto e = res.initExec();
        copyStruct(state, e);
      }
      break;
    default:
      return luaL_argerror(state, 3, "unknown resource type");
    }
    return 0;
  }

  const luaL_Reg mcmlib[] = {
    {"exec", execfunc},
    {"file", filefunc},
    {"hash", hashfunc},
    {"resource", resourcefunc},
    {NULL, NULL},
  };

  int openlib(lua_State* state) {
    luaL_newlibtable(state, mcmlib);
    lua_pushvalue(state, lua_upvalueindex(1));
    luaL_setfuncs(state, mcmlib, 1);

    lua_newtable(state);
    lua_createtable(state, 0, 1);  // new metatable
    pushResourceType(state, 0);
    lua_setfield(state, -2, resourceTypeMetaKey);  // metatable[resourceTypeMetaKey] = TOP
    lua_setmetatable(state, -2);  // pop metatable
    lua_setfield(state, -2, "noop");  // mcm.noop = TOP
    return 1;
  }

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
}  // namespace

Lua::Lua() {
  state = luaL_newstate();
  KJ_ASSERT_NONNULL(state);
  const luaL_Reg *lib;
  for (lib = loadedlibs; lib->func; lib++) {
    luaL_requiref(state, lib->name, lib->func, 1);
    lua_pop(state, 1);  // remove lib
  }

  luaL_getsubtable(state, LUA_REGISTRYINDEX, "_LOADED");
  lua_getfield(state, -1, "mcm");  // _LOADED["mcm"]
  if (!lua_toboolean(state, -1)) {  // package not already loaded?
    lua_pop(state, 1);  // remove field
    lua_pushlightuserdata(state, this);
    lua_pushcclosure(state, openlib, 1);
    lua_pushstring(state, "mcm");  // argument to open function
    lua_call(state, 1, 1);  // call openlib to open module
    lua_pushvalue(state, -1);  // make copy of module (call result)
    lua_setfield(state, -3, "mcm");  // _LOADED["mcm"] = module
  }
  lua_remove(state, -2);  // remove _LOADED table
  lua_setglobal(state, "mcm");  // _G["mcm"] = module

  KJ_ASSERT(lua_gettop(state) == 0, "all elements should have been popped");
}

void Lua::exec(kj::StringPtr fname) {
  if (luaL_loadfile(state, fname.cStr()) || lua_pcall(state, 0, 0, 0)) {
    auto errMsg = kj::heapString(luaStringPtr(state, -1));
    lua_pop(state, 1);
    KJ_FAIL_ASSERT(errMsg);
  }
}

Resource::Builder Lua::newResource() {
  auto orphan = scratch.getOrphanage().newOrphan<Resource>();
  auto builder = orphan.get();
  resources.add(kj::mv(orphan));
  return builder;
}

void Lua::finish(capnp::MessageBuilder& message) {
  auto catalog = message.initRoot<Catalog>();
  auto rlist = catalog.initResources(resources.size());
  // TODO(soon): sort
  for (size_t i = 0; i < resources.size(); i++) {
    rlist.setWithCaveats(i, resources[i].get());
  }
}

Lua::~Lua() {
  lua_close(state);
}

class Main {
public:
  Main(kj::ProcessContext& context): context(context) {}

  kj::MainBuilder::Validity processFile(kj::StringPtr src) {
    if (src.size() == 0) {
      return kj::str("empty source");
    }
    Lua l;
    l.exec(src);
    capnp::MallocMessageBuilder message;
    l.finish(message);
    kj::FdOutputStream out(STDOUT_FILENO);
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
