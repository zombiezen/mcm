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

MAIN_SRCS = ["luacat.c++", "version.h"]
TEST_GLOB = ["*-test.c++"]

cc_binary(
    name = "mcm-luacat",
    srcs = MAIN_SRCS,
    deps = [
        ":luacat",
        "//third_party/capnproto:kj",
    ],
)

genrule(
    name = "buildstamp",
    outs = ["version.h"],
    cmd = "$(location genversion.sh) > \"$@\"",
    tools = ["genversion.sh"],
    stamp = 1,
)

cc_library(
    name = "luacat",
    srcs = glob(
        [
            "*.c++",
            "*.h",
        ],
        exclude = MAIN_SRCS + TEST_GLOB,
    ),
    deps = [
        "//:catalog_cc",
        "//third_party/capnproto:capnp_lib",
        "//third_party/capnproto:kj",
        "//third_party/lua:lib",
        "@boringssl//:crypto",
    ],
)

cc_test(
    name = "tests",
    srcs = glob(TEST_GLOB),
    size = "small",
    deps = [
        ":luacat",
        ":testsuite",
        "//:catalog_cc",
        "@gtest//:gtest_main",
    ],
)

capnp_library(
    name = "testsuite_capnp",
    src = "testsuite.capnp",
    testonly = 1,
    data = glob(["testdata/*.lua"]),
    deps = [
        "//:catalog_capnp",
        "//third_party/capnproto:cc",
    ],
)

capnp_cc_library(
    name = "testsuite",
    lib = ":testsuite_capnp",
    basename = "testsuite.capnp",
    testonly = 1,
    deps = [
        "//:catalog_cc",
    ],
)
