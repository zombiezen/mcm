load("@io_bazel_rules_go//go:def.bzl", "go_prefix")

go_prefix("github.com/zombiezen/mcm")

capnp_library(
    name = "catalog_capnp",
    src = "catalog.capnp",
    deps = [
        "//third_party/capnproto:cc",
        "//third_party/golang/capnproto/std:go_capnp",
    ],
)

capnp_cc_library(
    name = "catalog_cc",
    lib = ":catalog_capnp",
    basename = "catalog.capnp",
    visibility = ["//visibility:public"],
)

capnp_go_library(
    name = "catalog",
    lib = ":catalog_capnp",
    visibility = ["//visibility:public"],
)
