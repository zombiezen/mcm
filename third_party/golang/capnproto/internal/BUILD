package(default_visibility = ["//third_party/golang/capnproto:__subpackages__"])
licenses(["notice"])

capnp_library(
    name = "aircraft_capnp",
    src = "aircraftlib/aircraft.capnp",
    testonly = 1,
    deps = [
        "//third_party/golang/capnproto/std:go_capnp",
    ],
)

capnp_go_library(
    name = "aircraftlib",
    lib = ":aircraft_capnp",
    testonly = 1,
)

go_default_library(
    name = "capnptool",
)

go_default_library(
    name = "fulfiller",
    test = 1,
    deps = [
        ":queue",
        "//third_party/golang/capnproto:go_default_library",
    ],
)

go_binary(
    name = "mktemplates",
    srcs = ["cmd/mktemplates/mktemplates.go"],
)

go_default_library(
    name = "nodemap",
    deps = [
        "//third_party/golang/capnproto:go_default_library",
        "//third_party/golang/capnproto:schemas",
        "//third_party/golang/capnproto/std/capnp:schema_bootstrap",
    ],
)

go_default_library(
    name = "packed",
    exclude = [
        "packed/discard_go14.go",
    ],
    test = 1,
)

go_default_library(
    name = "queue",
    test = 1,
)

filegroup(
    name = "aircraft_schema",
    srcs = ["aircraftlib/aircraft.capnp"],
)
