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
