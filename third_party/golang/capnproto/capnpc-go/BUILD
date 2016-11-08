licenses(["notice"])

go_binary(
    name = "capnpc-go",
    srcs = [":templates"] + glob(
        ["*.go"],
        exclude = ["*_test.go"],
    ),
    deps = [
        "//third_party/golang/capnproto:go_default_library",
        "//third_party/golang/capnproto/std/capnp:schema_bootstrap",
    ],
    visibility = ["//visibility:public"],
)

# go_test(
#     name = "test",
#     srcs = glob(["*_test.go"]),
#     library = ":capnpc-go",
#     deps = [
#         "//third_party/golang/capnproto:go_default_library",
#         "//third_party/golang/capnproto/encoding/text:go_default_library",
#         "//third_party/golang/capnproto/std/capnp:schema_bootstrap",
#     ],
#     data = glob(["testdata/*.out"]),
# )

genrule(
    name = "templates",
    srcs = glob(["templates/*"]),
    outs = ["templates.go"],
    cmd = "$(location //third_party/golang/capnproto/internal:mktemplates) $@ third_party/golang/capnproto/capnpc-go/templates",
    tools = ["//third_party/golang/capnproto/internal:mktemplates"],
)
