cc_binary(
    name = "mcm-luacat",
    srcs = glob(["*.c++"]),
    deps = [
        "//:catalog_cc",
        "//third_party/capnproto:capnp_lib",
        "//third_party/capnproto:kj",
        "//third_party/lua:lib",
        "@boringssl//:crypto",
    ],
)
