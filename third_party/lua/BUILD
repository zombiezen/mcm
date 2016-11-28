# Copyright 2016 The Minimal Configuration Manager Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

package(default_visibility = ["//visibility:public"])
licenses(["notice"])

LUA_SRCS = ["src/lua.c"]
LUAC_SRCS = ["src/luac.c"]
LUA_HDRS = [
    "src/lauxlib.h",
    "src/lua.h",
    "src/lua.hpp",
    "src/luaconf.h",
    "src/lualib.h",
]

cc_library(
    name = "lib",
    srcs = glob(
        [
            "src/*.c",
            "src/*.h",
        ],
        exclude = LUA_SRCS + LUAC_SRCS + LUA_HDRS,
    ),
    hdrs = LUA_HDRS,
    defines = ["LUA_USE_POSIX"],
    includes = ["src"],
    linkopts = ["-lm"],
)

cc_binary(
    name = "lua",
    srcs = LUA_SRCS,
    deps = [
        ":lib",
    ],
)

cc_binary(
    name = "luac",
    srcs = LUAC_SRCS,
    deps = [
        ":lib",
    ],
)
