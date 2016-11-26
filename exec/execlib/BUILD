package(default_visibility = ["//exec:__subpackages__"])

go_default_library(
    test = 1,
    deps = [
        "//:catalog",
        "//internal/depgraph:go_default_library",
        "//internal/system:go_default_library",
    ],
    test_deps = [
        "//:catalog",
        "//internal/system/fakesystem:go_default_library",
        "//third_party/golang/capnproto:go_default_library",
        "//third_party/golang/capnproto:pogs",
    ],
)
