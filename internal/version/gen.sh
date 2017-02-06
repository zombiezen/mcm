#!/bin/bash
# Copyright 2017 The Minimal Configuration Manager Authors
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

echo 'package version'

copyVal() {
  local statusFile="$1"
  local statusName="$2"
  local goName="$3"
  local val
  val="$(awk "/^${statusName}/ { print \$2 }" "$statusFile")"
  [[ $? -eq 0 ]] || return 1
  echo "const $goName = \"$val\""
}

copyVal bazel-out/stable-status.txt BUILD_EMBED_LABEL Label || exit 1
copyVal bazel-out/volatile-status.txt BUILD_SCM_REVISION SCMRevision || exit 1
copyVal bazel-out/volatile-status.txt BUILD_SCM_STATUS SCMStatus || exit 1
