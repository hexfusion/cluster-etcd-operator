#!/usr/bin/env bash

### Created by cluster-etcd-operator. DO NOT edit.

set -o errexit
set -o pipefail
set -o errtrace

# example
# ./cluster-restore.sh $path-to-backup

if [[ $EUID -ne 0 ]]; then
  echo "This script must be run as root"
  exit 1
fi

function source_required_dependency {
  local src_path="$1"
  if [ ! -f "${src_path}" ]; then
    echo "required dependencies not found, please ensure this script is run on a node with a functional etcd static pod"
    exit 1
  fi
  # shellcheck disable=SC1090
  source "${src_path}"
}

source_required_dependency /etc/kubernetes/static-pod-resources/etcd-certs/configmaps/etcd-scripts/etcd.env
source_required_dependency /etc/kubernetes/static-pod-resources/etcd-certs/configmaps/etcd-scripts/etcd-common-tools

function usage() {
  echo 'Path to the directory containing backup files is required: ./cluster-restore.sh <path-to-backup>'
  echo 'The backup directory is expected to be contain two files:'
  echo '        1. etcd snapshot'
  echo '        2. A copy of the Static POD resources at the time of backup'
  exit 1
}

# If the argument is not passed, or if it is not a directory, print usage and exit.
if [ "$1" == "" ] || [ ! -d "$1" ]; then
  usage
fi

function restore_static_pods() {
  local backup_file="$1"
  shift
  local static_pods=("$@")

  for pod_file_name in "${static_pods[@]}"; do
    backup_pod_path=$(tar -tvf "${backup_file}" "*${pod_file_name}" | awk '{ print $6 }') || true
    if [ -z "${backup_pod_path}" ]; then
      echo "${pod_file_name} does not exist in ${backup_file}"
      exit 1
    fi

    echo "starting ${pod_file_name}"
    tar -xvf "${backup_file}" --strip-components=2 -C "${MANIFEST_DIR}"/ "${backup_pod_path}"
  done
}

function wait_for_containers_to_stop() {
  local containers=("$@")

  for container_name in "${containers[@]}"; do
    echo "Waiting for container ${container_name} to stop"
    while [[ -n $(crictl ps --label io.kubernetes.container.name="${container_name}" -q) ]]; do
      echo -n "."
      sleep 1
    done
    echo "complete"
  done
}

BACKUP_DIR="$1"
# shellcheck disable=SC2012
BACKUP_FILE=$(ls -vd "${BACKUP_DIR}"/static_kuberesources*.tar.gz | tail -1) || true
# shellcheck disable=SC2012
SNAPSHOT_FILE=$(ls -vd "${BACKUP_DIR}"/snapshot*.db | tail -1) || true
STATIC_POD_LIST=("kube-apiserver-pod.yaml" "kube-controller-manager-pod.yaml" "kube-scheduler-pod.yaml")
STATIC_POD_CONTAINERS=("etcd" "etcdctl" "etcd-metrics" "kube-controller-manager" "kube-apiserver" "kube-scheduler")

if [ ! -f "${SNAPSHOT_FILE}" ]; then
  echo "etcd snapshot ${SNAPSHOT_FILE} does not exist"
  exit 1
fi

# Download etcdctl and check the snapshot status
dl_etcdctl
check_snapshot_status "${SNAPSHOT_FILE}"

# Move manifests and stop static pods
if [ ! -d "$MANIFEST_STOPPED_DIR" ]; then
  mkdir -p "$MANIFEST_STOPPED_DIR"
fi

# Move static pod manifests out of MANIFEST_DIR
for POD_FILE_NAME in "${STATIC_POD_LIST[@]}" etcd-pod.yaml; do
  echo "...stopping ${POD_FILE_NAME}"
  [ ! -f "${MANIFEST_DIR}/${POD_FILE_NAME}" ] && continue
  mv "${MANIFEST_DIR}/${POD_FILE_NAME}" "${MANIFEST_STOPPED_DIR}"
done

# wait for every static pod container to stop
wait_for_containers_to_stop "${STATIC_POD_CONTAINERS[@]}"

if [ ! -d "${ETCD_DATA_DIR_BACKUP}" ]; then
  mkdir -p "${ETCD_DATA_DIR_BACKUP}"
fi

# backup old data-dir
if [ -d "${ETCD_DATA_DIR}/member" ]; then
  if [ -d "${ETCD_DATA_DIR_BACKUP}/member" ]; then
    echo "removing previous backup ${ETCD_DATA_DIR_BACKUP}/member"
    rm -rf "${ETCD_DATA_DIR_BACKUP}"/member
  fi
  echo "Moving etcd data-dir ${ETCD_DATA_DIR}/member to ${ETCD_DATA_DIR_BACKUP}"
  mv "${ETCD_DATA_DIR}"/member "${ETCD_DATA_DIR_BACKUP}"/
fi

# Restore static pod resources
tar -C "${CONFIG_FILE_DIR}" -xzf "${BACKUP_FILE}" static-pod-resources

# Copy snapshot to backupdir
cp -p "${SNAPSHOT_FILE}" "${ETCD_DATA_DIR_BACKUP}"/snapshot.db

echo "starting restore-etcd static pod"
cp -p "${RESTORE_ETCD_POD_YAML}" "${MANIFEST_DIR}/etcd-pod.yaml"

# start remaining static pods
restore_static_pods "${BACKUP_FILE}" "${STATIC_POD_LIST[@]}"
