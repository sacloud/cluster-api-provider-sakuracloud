#!/bin/bash

# Copyright 2019 Kazumichi Yamamoto
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

# Change directories to the parent directory of the one in which this
# script is located.
cd "${WORKDIR:-$(dirname "${BASH_SOURCE[0]}")/..}"
BUILDDIR="${BUILDDIR:-.}"

OUT_DIR="${OUT_DIR:-}"
SRC_DIR="${BUILDDIR}"/examples

OVERWRITE=
CLUSTER_NAME="${CLUSTER_NAME:-caps-example}"
ENV_VAR_REQ=':?required'

CABPK_MANAGER_IMAGE="${CABPK_MANAGER_IMAGE:-us.gcr.io/k8s-artifacts-prod/capi-kubeadm/cluster-api-kubeadm-controller:v0.1.3}"
CAPI_MANAGER_IMAGE="${CAPI_MANAGER_IMAGE:-us.gcr.io/k8s-artifacts-prod/cluster-api/cluster-api-controller:v0.2.5}"
CAPS_MANAGER_IMAGE="${CAPS_MANAGER_IMAGE:-sacloud/cluster-api-provider-sakuracloud:latest}"
K8S_IMAGE_REPOSITORY="${K8S_IMAGE_REPOSITORY:-k8s.gcr.io}"

# Set the default log levels for the manager containers.
CABPK_MANAGER_LOG_LEVEL="${CABPK_MANAGER_LOG_LEVEL:-4}"
CAPI_MANAGER_LOG_LEVEL="${CAPI_MANAGER_LOG_LEVEL:-4}"
CAPS_MANAGER_LOG_LEVEL="${CAPV_MANAGER_LOG_LEVEL:-4}"

usage() {
  cat <<EOF
usage: ${0} [FLAGS]
  Generates input manifests for the Cluster API Provider for vSphere (CAPV)

FLAGS
  -b    bootstrapper manager image (default "${CABPK_MANAGER_IMAGE}")
  -B    bootstrapper manager log level (default "${CABPK_MANAGER_LOG_LEVEL}")
  -c    cluster name (default "${CLUSTER_NAME}")
  -d    disables required environment variables
  -f    force overwrite of existing files
  -h    prints this help screen
  -i    input directory (default ${SRC_DIR})
  -m    caps manager image (default "${CAPS_MANAGER_IMAGE}")
  -M    caps manager log level (default "${CAPS_MANAGER_LOG_LEVEL}")
  -r    kubernetes container image repository (default "${K8S_IMAGE_REPOSITORY}")
  -o    output directory (default ${OUT_DIR})
  -p    capi manager image (default "${CAPI_MANAGER_IMAGE}")
  -P    capi manager log level (default "${CAPI_MANAGER_LOG_LEVEL}")
EOF
}

while getopts ':b:B:c:dfhi:m:M:r:o:p:P' opt; do
  case "${opt}" in
  b)
    CABPK_MANAGER_IMAGE="${OPTARG}"
    ;;
  B)
    CABPK_MANAGER_LOG_LEVEL="${OPTARG}"
    ;;
  c)
    CLUSTER_NAME="${OPTARG}"
    ;;
  d)
    ENV_VAR_REQ=':-'
    ;;
  f)
    OVERWRITE=1
    ;;
  h)
    usage 1>&2; exit 1
    ;;
  i)
    SRC_DIR="${OPTARG}"
    ;;
  m)
    CAPS_MANAGER_IMAGE="${OPTARG}"
    ;;
  M)
    CAPS_MANAGER_LOG_LEVEL="${OPTARG}"
    ;;
  r)
    K8S_IMAGE_REPOSITORY="${OPTARG}"
    ;;
  o)
    OUT_DIR="${OPTARG}"
    ;;
  p)
    CAPI_MANAGER_IMAGE="${OPTARG}"
    ;;
  P)
    CAPI_MANAGER_LOG_LEVEL="${OPTARG}"
    ;;
  \?)
    { echo "invalid option: -${OPTARG}"; usage; } 1>&2; exit 1
    ;;
  :)
    echo "option -${OPTARG} requires an argument" 1>&2; exit 1
    ;;
  esac
done
shift $((OPTIND-1))

[ -n "${OUT_DIR}" ] || OUT_DIR="./out/${CLUSTER_NAME}"
mkdir -p "${OUT_DIR}"

# Load an envvars.txt file if one is found.
# shellcheck disable=SC1091
[ "${DOCKER_ENABLED-}" ] && [ -e "/envvars.txt" ] && source "/envvars.txt"

# Export the manager images and log levels for the different providers.
export CABPK_MANAGER_IMAGE CABPK_MANAGER_LOG_LEVEL
export CAPI_MANAGER_IMAGE CAPI_MANAGER_LOG_LEVEL
export CAPS_MANAGER_IMAGE CAPS_MANAGER_LOG_LEVEL

# Outputs
COMPONENTS_CLUSTER_API_GENERATED_FILE=${SRC_DIR}/provider-components/provider-components-cluster-api.yaml
COMPONENTS_KUBEADM_GENERATED_FILE=${SRC_DIR}/provider-components/provider-components-kubeadm.yaml
COMPONENTS_SAKURACLOUD_GENERATED_FILE=${SRC_DIR}/provider-components/provider-components-sakuracloud.yaml

ADDONS_GENERATED_FILE=${OUT_DIR}/addons.yaml
PROVIDER_COMPONENTS_GENERATED_FILE=${OUT_DIR}/provider-components.yaml
CLUSTER_GENERATED_FILE=${OUT_DIR}/cluster.yaml
CONTROLPLANE_GENERATED_FILE=${OUT_DIR}/controlplane.yaml
MACHINEDEPLOYMENT_GENERATED_FILE=${OUT_DIR}/machinedeployment.yaml

ok_file() {
  [ -f "${1}" ] || { echo "${1} is missing" 1>&2; exit 1; }
}

no_file() {
  [ ! -f "${1}" ] || { echo "${1} already exists, overwrite with -f" 1>&2; exit 1; }
}

# Remove the temporary provider components files.
for f in COMPONENTS_CLUSTER_API COMPONENTS_KUBEADM COMPONENTS_SAKURACLOUD; do \
  eval "rm -f \"\${${f}_GENERATED_FILE}\""
done

# Ensure that the actual outputs are only overwritten if the flag is provided.
for f in ADDONS PROVIDER_COMPONENTS CLUSTER CONTROLPLANE MACHINEDEPLOYMENT; do
  [ -n "${OVERWRITE}" ] || eval "no_file \"\${${f}_GENERATED_FILE}\""
done

require_if_defined() {
  while [ "${#}" -gt "0" ]; do
    eval "[ ! \"\${${1}+x}\" ] || ${1}=\"\${${1}${ENV_VAR_REQ}}\""
    shift
  done
}

require_if_defined CABPK_MANAGER_IMAGE \
                   CAPS_MANAGER_IMAGE \
                   SAKURACLOUD_ZONE \
                   SAKURACLOUD_SOURCE_ARCHIVE_NAME

# All variables used for yaml generation
EXPORTED_ENV_VARS=
record_and_export() {
  eval "EXPORTED_ENV_VARS=\"${EXPORTED_ENV_VARS} -e ${1}\"; \
        export ${1}=\"\${${1}${2}}\""
}
record_and_export CLUSTER_NAME                    ':-caps-example'
record_and_export SERVICE_CIDR                    ':-100.64.0.0/13'
record_and_export CLUSTER_CIDR                    ':-100.96.0.0/11'
record_and_export SERVICE_DOMAIN                  ':-cluster.local'
record_and_export CABPK_MANAGER_IMAGE             ':-'
record_and_export CAPS_MANAGER_IMAGE              ':-'
record_and_export K8S_IMAGE_REPOSITORY            ':-'
record_and_export SAKURACLOUD_ACCESS_TOKEN        "${ENV_VAR_REQ}"
record_and_export SAKURACLOUD_ACCESS_TOKEN_SECRET "${ENV_VAR_REQ}"
record_and_export SAKURACLOUD_ZONE                ":-'is1a'"
record_and_export SAKURACLOUD_SOURCE_ARCHIVE_NAME ":-capi-kubernetes-template"
record_and_export SAKURACLOUD_CONTROLPLANE_CPUS   ":-2"
record_and_export SAKURACLOUD_CONTROLPLANE_MEMORY ":-4"
record_and_export SAKURACLOUD_CONTROLPLANE_DISK   ":-20"
record_and_export SAKURACLOUD_MD_CPUS             ":-2"
record_and_export SAKURACLOUD_MD_MEMORY           ":-2"
record_and_export SAKURACLOUD_MD_DISK             ":-20"
record_and_export SSH_AUTHORIZED_KEY              ":-''"

[[ ${KUBERNETES_VERSION-} =~ ^v?[[:digit:]]+\.[[:digit:]]+\.[[:digit:]]+([\+\.\-](.+))?$ ]] || KUBERNETES_VERSION="1.15.4"
record_and_export KUBERNETES_VERSION ":-${KUBERNETES_VERSION}"

# Base64 encode the credentials and unset the plain-text values.
SAKURACLOUD_B64ENCODED_TOKEN="$(printf '%s' "${SAKURACLOUD_ACCESS_TOKEN}" | base64)"
SAKURACLOUD_B64ENCODED_SECRET="$(printf '%s' "${SAKURACLOUD_ACCESS_TOKEN_SECRET}" | base64)"
export SAKURACLOUD_B64ENCODED_TOKEN SAKURACLOUD_B64ENCODED_SECRET
unset SAKURACLOUD_ACCESS_TOKEN SAKURACLOUD_ACCESS_TOKEN_SECRET

envsubst() {
  python -c 'import os,sys;[sys.stdout.write(os.path.expandvars(l)) for l in sys.stdin]'
}

# Generate the addons file.
envsubst >"${ADDONS_GENERATED_FILE}" <"${SRC_DIR}/addons.yaml"
echo "Generated ${ADDONS_GENERATED_FILE}"

# Generate cluster resources.
kustomize build "${SRC_DIR}/cluster" | envsubst >"${CLUSTER_GENERATED_FILE}"
echo "Generated ${CLUSTER_GENERATED_FILE}"

# Generate controlplane resources.
kustomize build "${SRC_DIR}/controlplane" | envsubst >"${CONTROLPLANE_GENERATED_FILE}"
echo "Generated ${CONTROLPLANE_GENERATED_FILE}"

# Generate machinedeployment resources.
kustomize build "${SRC_DIR}/machinedeployment" | envsubst >"${MACHINEDEPLOYMENT_GENERATED_FILE}"
echo "Generated ${MACHINEDEPLOYMENT_GENERATED_FILE}"

# Generate Cluster API provider components file.
kustomize build "github.com/kubernetes-sigs/cluster-api/config/default/?ref=v0.2.5" >"${COMPONENTS_CLUSTER_API_GENERATED_FILE}"
echo "Generated ${COMPONENTS_CLUSTER_API_GENERATED_FILE}"

# Generate Kubeadm Bootstrap Provider components file.
kustomize build "github.com/kubernetes-sigs/cluster-api-bootstrap-provider-kubeadm/config/default/?ref=v0.1.3" >"${COMPONENTS_KUBEADM_GENERATED_FILE}"
echo "Generated ${COMPONENTS_KUBEADM_GENERATED_FILE}"

# Generate VSphere Infrastructure Provider components file.
kustomize build "${SRC_DIR}/../config/default" | envsubst >"${COMPONENTS_SAKURACLOUD_GENERATED_FILE}"
echo "Generated ${COMPONENTS_SAKURACLOUD_GENERATED_FILE}"

# Generate a single provider components file.
kustomize build "${SRC_DIR}/provider-components" | envsubst >"${PROVIDER_COMPONENTS_GENERATED_FILE}"
echo "Generated ${PROVIDER_COMPONENTS_GENERATED_FILE}"
echo "WARNING: ${PROVIDER_COMPONENTS_GENERATED_FILE} includes SakuraCloud credentials"

# If running in Docker then ensure the contents of the OUT_DIR have the
# the same owner as the volume mounted to the /out directory.
[ "${DOCKER_ENABLED-}" ] && chown -R "$(stat -c '%u:%g' /out)" "${OUT_DIR}"
