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

package(
    default_testonly = 1,
    default_visibility = ["//visibility:public"],
)
licenses(["notice"])  # BSD

cc_library(
    name = "gtest",
    srcs = glob(
        [
            "googletest/src/*.cc",
            "googletest/src/*.h",
        ],
        exclude = [
            "googletest/src/gtest-all.cc",
            "googletest/src/gtest_main.cc",
        ],
    ),
    hdrs = glob([
        "googletest/include/**/*.h",
    ]),
    includes = [
        "googletest/include",
    ],
    copts = [
        "-Iexternal/gtest/googletest",
    ],
    linkopts = ["-pthread"],
    testonly = 1,
)

cc_library(
    name = "gtest_main",
    srcs = ["googletest/src/gtest_main.cc"],
    deps = [":gtest"],
)
