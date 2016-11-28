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

package(default_visibility = ["//visibility:public"])
licenses(["notice"])

INTEGRATION_TEST_SRCS = [
    "integration_test.go",
    "integrationutil_test.go",
]

go_default_library(
    test = 1,
    exclude = INTEGRATION_TEST_SRCS + [
        "example_test.go",
    ],
    deps = [
        "//third_party/golang/capnproto/internal:packed",
        "//third_party/golang/rbtree:go_default_library",
    ],
)

go_test(
    name = "integration_test",
    srcs = INTEGRATION_TEST_SRCS,
    size = "small",
    data = [
        "//third_party/capnproto:capnp",
        "//third_party/golang/capnproto/internal:aircraft_schema",
    ],
    deps = [
        ":go_default_library",
        "//third_party/golang/capnproto/internal:aircraftlib",
        "//third_party/golang/capnproto/internal:capnptool",
    ],
)

go_default_library(
    name = "pogs",
    exclude = [
        "pogs/example_test.go",
    ],
    deps = [
        ":go_default_library",
        "//third_party/golang/capnproto/internal:nodemap",
        "//third_party/golang/capnproto/std/capnp:schema_bootstrap",
    ],
    # test = 1,
    # test_deps = [
    #     ":go_default_library",
    #     "//third_party/golang/capnproto/internal:aircraftlib",
    #     "//third_party/golang/debug:pretty",
    # ],
)

go_default_library(
    name = "schemas",
    test = 1,
    test_separate = 1,
    deps = [
        "//third_party/golang/capnproto/internal:packed",
    ],
    test_deps = [
        ":go_default_library",
        ":schemas",
        "//third_party/golang/capnproto/std:go_capnp_go",
        "//third_party/golang/capnproto/std/capnp:schema_bootstrap",
    ],
)

go_default_library(
    name = "server",
    test = 1,
    test_separate = 1,
    deps = [
        ":go_default_library",
        "//third_party/golang/capnproto/internal:fulfiller",
    ],
    test_deps = [
        ":server",
        "//third_party/golang/capnproto/internal:aircraftlib",
    ],
)
