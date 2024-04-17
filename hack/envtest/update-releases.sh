#!/usr/bin/env bash
# Copyright 2024 The Kubernetes Authors.
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

set -o errexit
set -o nounset
set -o pipefail
set -x

ROOT=$(dirname "${BASH_SOURCE[0]}")/../..

if [ -z "${KUBERNETES_VERSION}" ]; then
  echo "Missing KUBERNETES_VERSION environment variable"
  exit 1
fi

# Create the releases.yaml file in hack/envtest if it does not exist
if [ ! -f "${ROOT}"/envtest-releases.yaml ]; then
  echo "releases:" > "${ROOT}"/envtest-releases.yaml
fi

# Add the newly built Kubernetes version to the releases.yaml file with yq as an object key under releases
yq eval ".releases += {\"${KUBERNETES_VERSION}\": {}}" -i "${ROOT}"/envtest-releases.yaml

for file in "${ROOT}"/out/*.tar.gz; do
  file_name=$(basename "${file}")
  file_hash=$(awk '{ print $1 }' < "${file}.sha512")
  self_link=https://github.com/kubernetes-sigs/controller-tools/releases/download/envtest-${KUBERNETES_VERSION}/${file_name}

  yq eval \
    ".releases[\"${KUBERNETES_VERSION}\"] += {\"${file_name}\": {\"hash\": \"${file_hash}\", \"selfLink\": \"${self_link}\"}}" \
    -i "${ROOT}"/envtest-releases.yaml
done
