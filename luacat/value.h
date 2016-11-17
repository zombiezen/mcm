#include <unistd.h>
#include <stdint.h>

#include "kj/common.h"
#include "kj/debug.h"
#include "kj/memory.h"
#include "kj/string.h"

extern "C" {
#include "lua.h"
}

namespace mcm {

namespace luacat {

class Id {
public:
  inline Id(uint64_t v, kj::StringPtr c) : value(v), comment(kj::heapString(c)) {}
  KJ_DISALLOW_COPY(Id);

  inline uint64_t getValue() const { return value; }
  inline kj::StringPtr getComment() const { return comment; }

private:
  uint64_t value;
  kj::String comment;
};

void pushResourceType(lua_State* state, uint64_t rt);
kj::Maybe<uint64_t> getResourceType(lua_State* state, int index);

void pushId(lua_State* state, kj::Own<const Id> id);
kj::Maybe<const Id&> getId(lua_State* state, int index);

}  // namespace luacat
}  // namespace mcm
