package(default_visibility = ["//visibility:public"])
licenses(["notice"])

go_default_library(
    test = 1,
    deps = [
        "//third_party/golang/capnproto:go_default_library",
        "//third_party/golang/capnproto:schemas",
        "//third_party/golang/capnproto/internal:nodemap",
        "//third_party/golang/capnproto/std/capnp:schema_bootstrap",
    ],
    test_deps = [
        "//third_party/golang/capnproto:go_default_library",
        "//third_party/golang/capnproto:schemas",
        "//third_party/golang/capnproto/std/capnp:schema_bootstrap",
    ],
)
