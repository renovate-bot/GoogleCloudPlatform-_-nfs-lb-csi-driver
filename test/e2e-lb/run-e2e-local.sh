#!/bin/bash
# Copyright 2024 Google LLC
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

set -o xtrace
set -o nounset
set -o errexit

readonly PKGDIR="$( dirname -- "$0"; )/../.."
readonly gcp_project=${PROJECT:-}
readonly ginkgo_focus="${E2E_TEST_FOCUS:-}"
readonly ginkgo_skip="${E2E_TEST_SKIP:-}"
readonly ginkgo_procs="${E2E_TEST_GINKGO_PROCS:-1}"
readonly ginkgo_timeout="${E2E_TEST_GINKGO_TIMEOUT:-1h}"
readonly ginkgo_flake_attempts="${E2E_TEST_GINKGO_FLAKE_ATTEMPTS:-1}"
readonly nfs_server_count=${E2E_TEST_NFS_SERVER_COUNT:-0}
readonly build_csi_driver_image=${E2E_BUILD_CSI_DRIVER_IMAGE:-true}
# readonly ip_list_file=${IP_LIST_FILE:-}
readonly install_csi_driver=${E2E_INSTALL_CSI_DRIVER:-false}
readonly destroy_csi_driver=${E2E_DESTROY_CSI_DRIVER:-false}

# Initialize ginkgo.
export PATH=${PATH}:$(go env GOPATH)/bin
go install github.com/onsi/ginkgo/v2/ginkgo@v2.17.1

# Build e2e-test CLI
go build -mod=vendor -o ${PKGDIR}/bin/e2e-lb-test ./test/e2e-lb
chmod +x ${PKGDIR}/bin/e2e-lb-test


if [ "$build_csi_driver_image" == "true" ]; then
    # build CSI driver container images
    make build-nfs-csi-image-and-push
fi

#Install the CSI driver
# PROJECT=${PROJECT} IP_LIST_FILE=${ip_list_file} make helm-csi-install

base_cmd="${PKGDIR}/bin/e2e-lb-test \
            --pkg-dir=${PKGDIR} \
            --ginkgo-focus=${ginkgo_focus} \
            --ginkgo-skip=${ginkgo_skip} \
            --ginkgo-procs=${ginkgo_procs} \
            --ginkgo-timeout=${ginkgo_timeout} \
            --ginkgo-flake-attempts=${ginkgo_flake_attempts} \
            --nfs-server-count=${nfs_server_count} \
            --gcp-project=${gcp_project} \
            --install-csi-driver=${install_csi_driver} \
            --destroy-csi-driver=${destroy_csi_driver}"


eval "$base_cmd"