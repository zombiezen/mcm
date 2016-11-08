package(default_visibility = ["//visibility:public"])
licenses(["notice"])

go_library(
    name = "schema_bootstrap",
    srcs = ["schema/schema.capnp.go"],
    deps = [
        "//third_party/golang/capnproto:go_default_library",
    ],
    visibility = [
        "//third_party/golang/capnproto:__subpackages__",
    ],
)
