package(default_visibility = ["//exec:__subpackages__"])

go_default_library(
    # TODO(windows): use select to enable this
    exclude = ["windows.go"],
    deps = [
        "//:catalog",
        "//internal/depgraph:go_default_library",
    ],
)
