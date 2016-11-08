package(default_visibility = ["//visibility:public"])
licenses(["notice"])  # MIT

cc_binary(
    name = "capnp",
    srcs = [
        "c++/src/capnp/compiler/capnp.c++",
        "c++/src/capnp/compiler/module-loader.c++",
        "c++/src/capnp/compiler/module-loader.h",
    ],
    deps = [
        ":capnp_lib",
        ":capnpc_lib",
        ":kj",
    ],
    includes = ["c++/src"],
)

cc_binary(
    name = "capnpc-c++",
    srcs = [
        "c++/src/capnp/compiler/capnpc-c++.c++",
    ],
    deps = [
        ":capnp_lib",
        ":kj",
    ],
    includes = ["c++/src"],
)

cc_binary(
    name = "capnpc-capnp",
    srcs = [
        "c++/src/capnp/compiler/capnpc-capnp.c++",
    ],
    deps = [
        ":capnp_lib",
        ":kj",
    ],
    includes = ["c++/src"],
)

CAPNPC_LIB_SRCS = [
    "c++/src/capnp/compiler/md5.c++",
    "c++/src/capnp/compiler/error-reporter.c++",
    "c++/src/capnp/compiler/lexer.capnp.c++",
    "c++/src/capnp/compiler/lexer.c++",
    "c++/src/capnp/compiler/grammar.capnp.c++",
    "c++/src/capnp/compiler/parser.c++",
    "c++/src/capnp/compiler/node-translator.c++",
    "c++/src/capnp/compiler/compiler.c++",
    "c++/src/capnp/schema-parser.c++",
]

cc_library(
    name = "kj",
    srcs = [
        "c++/src/kj/arena.c++",
        "c++/src/kj/array.c++",
        "c++/src/kj/async.c++",
        "c++/src/kj/async-io.c++",
        "c++/src/kj/async-unix.c++",
        "c++/src/kj/common.c++",
        "c++/src/kj/debug.c++",
        "c++/src/kj/exception.c++",
        "c++/src/kj/io.c++",
        "c++/src/kj/main.c++",
        "c++/src/kj/memory.c++",
        "c++/src/kj/mutex.c++",
        "c++/src/kj/parse/char.c++",
        "c++/src/kj/refcount.c++",
        "c++/src/kj/string-tree.c++",
        "c++/src/kj/string.c++",
        "c++/src/kj/thread.c++",
        "c++/src/kj/time.c++",
        "c++/src/kj/units.c++",
    ],
    hdrs = [
        "c++/src/kj/arena.h",
        "c++/src/kj/array.h",
        "c++/src/kj/async.h",
        "c++/src/kj/async-inl.h",
        "c++/src/kj/async-io.h",
        "c++/src/kj/async-prelude.h",
        "c++/src/kj/async-unix.h",
        "c++/src/kj/common.h",
        "c++/src/kj/debug.h",
        "c++/src/kj/exception.h",
        "c++/src/kj/function.h",
        "c++/src/kj/io.h",
        "c++/src/kj/main.h",
        "c++/src/kj/memory.h",
        "c++/src/kj/miniposix.h",
        "c++/src/kj/mutex.h",
        "c++/src/kj/one-of.h",
        "c++/src/kj/parse/char.h",
        "c++/src/kj/parse/common.h",
        "c++/src/kj/refcount.h",
        "c++/src/kj/std/iostream.h",
        "c++/src/kj/string-tree.h",
        "c++/src/kj/string.h",
        "c++/src/kj/thread.h",
        "c++/src/kj/threadlocal.h",
        "c++/src/kj/time.h",
        "c++/src/kj/tuple.h",
        "c++/src/kj/units.h",
        "c++/src/kj/vector.h",
    ],
    includes = ["c++/src"],
    linkopts = [
        "-pthread",
    ],
)

cc_library(
    name = "capnp_lib",
    srcs = [
        "c++/src/capnp/any.c++",
        "c++/src/capnp/arena.c++",
        "c++/src/capnp/blob.c++",
        "c++/src/capnp/capability.c++",
        "c++/src/capnp/dynamic.c++",
        "c++/src/capnp/layout.c++",
        "c++/src/capnp/list.c++",
        "c++/src/capnp/message.c++",
        "c++/src/capnp/schema-loader.c++",
        "c++/src/capnp/schema.c++",
        "c++/src/capnp/schema.capnp.c++",
        "c++/src/capnp/serialize-packed.c++",
        "c++/src/capnp/serialize.c++",
        "c++/src/capnp/serialize-async.c++",
        "c++/src/capnp/stringify.c++",
    ],
    hdrs = [
        "c++/src/capnp/any.h",
        "c++/src/capnp/arena.h",
        "c++/src/capnp/blob.h",
        "c++/src/capnp/common.h",
        "c++/src/capnp/capability.h",
        "c++/src/capnp/dynamic.h",
        "c++/src/capnp/endian.h",
        "c++/src/capnp/generated-header-support.h",
        "c++/src/capnp/layout.h",
        "c++/src/capnp/list.h",
        "c++/src/capnp/message.h",
        "c++/src/capnp/orphan.h",
        "c++/src/capnp/pointer-helpers.h",
        "c++/src/capnp/pretty-print.h",
        "c++/src/capnp/schema-loader.h",
        "c++/src/capnp/schema.capnp.h",
        "c++/src/capnp/schema.h",
        "c++/src/capnp/schema-lite.h",
        "c++/src/capnp/serialize-async.h",
        "c++/src/capnp/serialize-packed.h",
        "c++/src/capnp/serialize.h",
    ],
    deps = [
        ":kj",
    ],
    includes = ["c++/src"],
)

cc_library(
    name = "capnpc_lib",
    srcs = CAPNPC_LIB_SRCS,
    hdrs = [name.replace('.c++', '.h') for name in CAPNPC_LIB_SRCS],
    deps = [
        ":capnp_lib",
        ":kj",
    ],
    includes = ["c++/src"],
)

cc_library(
    name = "rpc_lib",
    srcs = [
        "c++/src/capnp/ez-rpc.c++",
        "c++/src/capnp/rpc-twoparty.c++",
        "c++/src/capnp/rpc.c++",
    ],
    hdrs = [
        "c++/src/capnp/ez-rpc.h",
        "c++/src/capnp/rpc-prelude.h",
        "c++/src/capnp/rpc-twoparty.h",
        "c++/src/capnp/rpc.h",
    ],
    deps = [
        ":capnp_lib",
        ":kj",
        ":persistent_cc",
        ":rpc_cc",
        ":rpc_twoparty_cc",
    ],
    includes = ["c++/src"],
)

# C++ annotations schema
capnp_library(
    name = "cc",
    src = "c++/src/capnp/c++.capnp",
    deps = [
        "//third_party/golang/capnproto/std:go_capnp",
    ],
)

capnp_library(
    name = "rpc",
    src = "c++/src/capnp/rpc.capnp",
    deps = [
        ":cc",
        "//third_party/golang/capnproto/std:go_capnp",
    ],
)

capnp_cc_library(
    name = "rpc_cc",
    lib = ":rpc",
    basename = "c++/src/capnp/rpc.capnp",
)

capnp_go_library(
    name = "rpc_go",
    lib = ":rpc",
)

capnp_library(
    name = "rpc_twoparty",
    src = "c++/src/capnp/rpc-twoparty.capnp",
    deps = [
        ":cc",
        "//third_party/golang/capnproto/std:go_capnp",
    ],
)

capnp_cc_library(
    name = "rpc_twoparty_cc",
    lib = ":rpc_twoparty",
    basename = "c++/src/capnp/rpc-twoparty.capnp",
)

capnp_library(
    name = "persistent",
    src = "c++/src/capnp/persistent.capnp",
    deps = [
        ":cc",
        "//third_party/golang/capnproto/std:go_capnp",
    ],
)

capnp_cc_library(
    name = "persistent_cc",
    lib = ":persistent",
    basename = "c++/src/capnp/persistent.capnp",
)
