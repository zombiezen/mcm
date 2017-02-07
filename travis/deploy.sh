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

build_label="$1"
gcs_out="${2}$(go env GOOS)_$(go env GOARCH)${3}"

echostep() {
  echo "$*" 1>&2
  "$@"
}

# Install gcloud & gsutil
if [[ ! -d "$HOME/google-cloud-sdk" ]]; then
  echostep curl 'https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-142.0.0-linux-x86_64.tar.gz' > /tmp/google-cloud-sdk.tar.gz || exit 1
  echo 'c05f649623b7a8696923f4c003bd95decb591f6e /tmp/google-cloud-sdk.tar.gz' | echostep sha1sum -c || exit 1
  echostep tar zxf /tmp/google-cloud-sdk.tar.gz -C "$HOME" || exit 1
fi
echostep $HOME/google-cloud-sdk/bin/gcloud --quiet components install gsutil || exit 1

# gcloud init
echostep $HOME/google-cloud-sdk/bin/gcloud config set disable_prompts True || exit 1
echostep $HOME/google-cloud-sdk/bin/gcloud config set project mcm-releases || exit 1
if [[ ! -f travis/service-account.json ]]; then
  echo "decrypt travis/service-account.json" 1>&2
  openssl aes-256-cbc \
    -K $encrypted_0c4c8f78bd1d_key -iv $encrypted_0c4c8f78bd1d_iv \
    -in travis/service-account.json.enc \
    -out travis/service-account.json \
    -d || exit 1
fi
echostep $HOME/google-cloud-sdk/bin/gcloud auth activate-service-account --key-file=travis/service-account.json || exit 1

# Build and deploy
echostep ./bazel --bazelrc=travis/bazelrc build -c opt --stamp --embed_label="$build_label" \
  //dot:mcm-dot //exec:mcm-exec //luacat:mcm-luacat //shellify:mcm-shellify || exit 1
echostep zip -j travis/build.zip \
  bazel-bin/dot/mcm-dot \
  bazel-bin/exec/mcm-exec \
  bazel-bin/luacat/mcm-luacat \
  bazel-bin/shellify/mcm-shellify || exit 1
echostep $HOME/google-cloud-sdk/bin/gsutil cp -n travis/build.zip "$gcs_out"
gsutil_result=$?
echostep rm -f travis/build.zip || exit 1
[[ $gsutil_result -eq 0 ]] || exit 1
