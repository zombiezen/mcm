# Copyright 2016 The Minimal Configuration Manager Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

load("@io_bazel_rules_go//go:def.bzl", "go_prefix")

go_prefix("github.com/zombiezen/mcm")

capnp_library(
    name = "catalog_capnp",
    src = "catalog.capnp",
    deps = [
        "//third_party/capnproto:cc",
        "//third_party/golang/capnproto/std:go_capnp",
    ],
    visibility = ["//visibility:public"],
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
