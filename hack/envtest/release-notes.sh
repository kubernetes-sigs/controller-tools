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

# This script is used to build the test binaries for the envtest package.
if [ -z "${KUBERNETES_VERSION}" ]; then
  echo "Missing KUBERNETES_VERSION environment variable"
  exit 1
fi

# For each file in out/*.tar.gz, build a out/release-notes.md files containing a table
# with every file and their respective hash

# Create the release notes file
echo -e "# Envtest Kubernetes ${KUBERNETES_VERSION} Binaries\n" > out/release-notes.md

# Create the table header
echo "filename | sha512 hash" >> out/release-notes.md
echo "-------- | -----------" >> out/release-notes.md

for file in "${ROOT}"/out/*.tar.gz; do
  # Get the file name
  file_name=$(basename "${file}")

  # Get the hash of the file
  file_hash=$(awk '{ print $1 }' < "${file}.sha512")

  # Add the file and hash to the release notes
  echo "| [${file_name}](https://github.com/kubernetes-sigs/controller-tools/releases/download/envtest-${KUBERNETES_VERSION}/${file_name}) | ${file_hash} |" >> out/release-notes.md
done

# Close the table
echo "" >> "${ROOT}"/out/release-notes.md
