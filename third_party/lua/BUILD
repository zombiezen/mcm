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
