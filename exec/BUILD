package(default_visibility = ["//godfather:__subpackages__"])

go_binary(
    name = "mcm-exec",
    srcs = glob(["*.go"]),
    deps = [
        "//:catalog",
        "//internal/depgraph:go_default_library",
        "//third_party/golang/capnproto:go_default_library",
    ],
)
