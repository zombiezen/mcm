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

package(default_visibility = ["//shellify:__subpackages__"])

X_TEST_SRCS = [
    "integration_test.go",
]

go_default_library(
    exclude = X_TEST_SRCS,
    test = 1,
    deps = [
        "//:catalog",
        "//internal/depgraph:go_default_library",
    ],
)

go_test(
    name = "go_default_library_x_test",
    srcs = X_TEST_SRCS,
    size = "medium",
    deps = [
        ":go_default_library",
        "//:catalog",
        "//internal/catpogs:go_default_library",
    ],
)
