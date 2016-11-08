package(default_visibility = ["//visibility:public"])
licenses(["notice"])

go_default_library(
    test = 1,
    test_separate = 1,
    deps = [
        "//third_party/capnproto:rpc_go",
        "//third_party/golang/capnproto:go_default_library",
        "//third_party/golang/capnproto/internal:fulfiller",
        "//third_party/golang/capnproto/internal:queue",
        "//third_party/golang/capnproto/rpc/internal:refcount",
    ],
    test_deps = [
        ":go_default_library",
        "//third_party/capnproto:rpc_go",
        "//third_party/golang/capnproto:go_default_library",
        "//third_party/golang/capnproto:server",
        "//third_party/golang/capnproto/rpc/internal:logtransport",
        "//third_party/golang/capnproto/rpc/internal:pipetransport",
        "//third_party/golang/capnproto/rpc/internal:testcapnp",
    ],
)
