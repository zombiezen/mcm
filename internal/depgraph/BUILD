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

package(default_visibility = ["//:__subpackages__"])

go_default_library(
    test = 1,
    deps = [
        "//:catalog",
    ],
    test_deps = [
        "//:catalog",
        "//third_party/golang/capnproto:go_default_library",
        "//third_party/golang/capnproto:pogs",
    ],
)
