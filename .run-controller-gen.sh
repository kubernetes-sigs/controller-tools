#!/usr/bin/env bash

# This builds and runs controller-gen in a particular context
# it's the equivalent of `go run sigs.k8s.io/controller-tools/cmd/controller-gen`
# if you could somehow do that without modifying your go.mod.

set -o errexit
set -o nounset
set -o pipefail

readlink=$(command -v readlink)

check_readlink() {
    local test_file="$(mktemp)"
    trap "rm -f $test_file" EXIT
    if ! ${readlink} -f "$test_file" &>/dev/null; then
        if [[ "${OSTYPE}" == "darwin"* ]]; then
            if command -v greadlink; then
                readlink=$(command -v greadlink)
                return
            fi
        fi
        echo "you're probably on OSX.  Please install gnu readlink -- otherwise you're missing the most useful readlink flag."
        exit 1
    fi
}
current_dir=$(pwd)
check_readlink
cd $(dirname $(${readlink} -f ${BASH_SOURCE[0]}))
go run -v -exec "./.run-in.sh ${current_dir} " ./cmd/controller-gen $@
