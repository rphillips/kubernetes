#!/usr/bin/env bash

set -eou pipefail

export ENABLE_SELFHOSTED_API="true"
export BOOTSTRAP_CHECKPOINT_PATH=$(mktemp -d)

echo "Enable Self Hosted API Server: ${ENABLE_SELFHOSTED_API}"
echo "Bootstrap Checkpoint Path: ${BOOTSTRAP_CHECKPOINT_PATH}"
echo

cleanup() {
  sudo rm -rf ${BOOTSTRAP_CHECKPOINT_PATH}
  sudo rm -rf /tmp/kube-*.log
  docker stop $(docker ps -a -q)
  docker rm $(docker ps -a -q)
}
trap cleanup EXIT

sudo ENABLE_SELFHOSTED_API=${ENABLE_SELFHOSTED_API} \
     BOOTSTRAP_CHECKPOINT_PATH=${BOOTSTRAP_CHECKPOINT_PATH} \
     PATH="$PWD:$PATH" \
     hack/local-up-cluster.sh
