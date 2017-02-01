#!/bin/bash
LUA_PATH="$TEST_SRCDIR/com_zombiezen_mcm/luacat/lib/?.lua" \
    exec "$TEST_SRCDIR/com_zombiezen_mcm/third_party/lua/lua" \
        "$TEST_SRCDIR/com_zombiezen_mcm/luacat/lib/configs_test.lua"
