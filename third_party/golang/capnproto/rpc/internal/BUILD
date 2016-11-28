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

package(default_visibility = ["//third_party/golang/capnproto/rpc:__subpackages__"])
licenses(["notice"])

go_default_library(
    name = "logtransport",
    testonly = 1,
    deps = [
        ":logutil",
        "//third_party/capnproto:rpc_go",
        "//third_party/golang/capnproto/encoding/text:go_default_library",
        "//third_party/golang/capnproto/rpc:go_default_library",
    ],
)

go_default_library(
    name = "logutil",
)

go_default_library(
    name = "pipetransport",
    testonly = 1,
    deps = [
        "//third_party/capnproto:rpc_go",
        "//third_party/golang/capnproto:go_default_library",
        "//third_party/golang/capnproto/rpc:go_default_library",
    ],
)

go_default_library(
    name = "refcount",
    test = 1,
    deps = [
        "//third_party/golang/capnproto:go_default_library",
    ],
)

capnp_library(
    name = "testcapnp_schema",
    src = "testcapnp/test.capnp",
    testonly = 1,
    deps = [
        "//third_party/golang/capnproto/std:go_capnp",
    ],
)

capnp_go_library(
    name = "testcapnp",
    lib = ":testcapnp_schema",
    testonly = 1,
)
