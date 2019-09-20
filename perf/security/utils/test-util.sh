#!/bin/bash

# Copyright 2019 Istio Authors
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

# This script help to dump information from a perf test deployment for debugging.
# It offers the following options:
# - proxyconfig: Dump all sidecar proxy configs.
# - proxylog: Dump all sidecar proxy logs.
# - setproxyloglevel: Change all sidecar proxy log level
# - proxycert: Dump all sidecar proxy /certs endpoint

NAMESPACE=${NAMESPACE:?"specify the namespace for running the test"}
CLUSTER=${CLUSTER:?"specify the cluster for running the test"}

LOGLEVEL=${2:-info}

proxyconfig() {
  # shellcheck disable=SC2086
  sleep_pods=$(kubectl get pods -n ${NAMESPACE} -o go-template --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' -l app=sleep --cluster ${CLUSTER})
  # shellcheck disable=SC2086
  httpbin_pods=$(kubectl get pods -n ${NAMESPACE} -o go-template --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' -l app=httpbin --cluster ${CLUSTER})

  pods=()

  while read -r line; do
    pods+=("$line")
  done <<< "${sleep_pods}"
  while read -r line; do
    pods+=("$line")
  done <<< "${httpbin_pods}"

  if [ ${#pods[@]} = 0 ]; then
    echo "no pods found!"
    exit
  fi

  configdir="proxy-config-$(date +"%Y%m%d%H%M%S%3N")"
  mkdir "/tmp/${configdir}"
  for pod in "${pods[@]}"
  do
    configpath="/tmp/${configdir}/${pod}-proxy.config"
    touch "${configpath}"
    echo "Dump istio-proxy config from pod ${pod} into ${configpath}"
    # shellcheck disable=SC2086
    kubectl exec -it -n ${NAMESPACE} "${pod}" -c istio-proxy --cluster ${CLUSTER} -- curl 127.0.0.1:15000/config_dump  > "${configpath}"
  done
}

proxylog() {
  # shellcheck disable=SC2086
  sleep_pods=$(kubectl get pods -n ${NAMESPACE} -o go-template --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' -l app=sleep --cluster ${CLUSTER})
  # shellcheck disable=SC2086
  httpbin_pods=$(kubectl get pods -n ${NAMESPACE} -o go-template --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' -l app=httpbin --cluster ${CLUSTER})

  pods=()

  while read -r line; do
    pods+=("$line")
  done <<< "${sleep_pods}"
  while read -r line; do
    pods+=("$line")
  done <<< "${httpbin_pods}"

  if [ ${#pods[@]} = 0 ]; then
    echo "no pods found!"
    exit
  fi

  logdir="proxy-log-$(date +"%Y%m%d%H%M%S%3N")"
  mkdir "/tmp/${logdir}"
  for pod in "${pods[@]}"
  do
    logpath="/tmp/${logdir}/${pod}.log"
    touch "${logpath}"
    echo "Dump istio-proxy logs from pod ${pod} into ${logpath}"
    # shellcheck disable=SC2086
    kubectl logs -n ${NAMESPACE} "${pod}" -c istio-proxy --cluster ${CLUSTER} > "${logpath}"
  done
}

setproxyloglevel() {
  # shellcheck disable=SC2086
  sleep_pods=$(kubectl get pods -n ${NAMESPACE} -o go-template --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' -l app=sleep --cluster ${CLUSTER})
  # shellcheck disable=SC2086
  httpbin_pods=$(kubectl get pods -n ${NAMESPACE} -o go-template --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' -l app=httpbin --cluster ${CLUSTER})

  pods=()

  while read -r line; do
    pods+=("$line")
  done <<< "${sleep_pods}"
  while read -r line; do
    pods+=("$line")
  done <<< "${httpbin_pods}"

  if [ ${#pods[@]} = 0 ]; then
    echo "no pods found!"
    exit
  fi
  for pod in "${pods[@]}"
  do
    echo "Set istio-proxy log level in pod ${pod} to ${LOGLEVEL}"
    # shellcheck disable=SC2086
    kubectl exec -it -n ${NAMESPACE} "${pod}" -c istio-proxy --cluster ${CLUSTER} -- curl -X POST 127.0.0.1:15000/logging?level="${LOGLEVEL}"
  done
}

proxycert() {
  # shellcheck disable=SC2086
  sleep_pods=$(kubectl get pods -n ${NAMESPACE} -o go-template --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' -l app=sleep --cluster ${CLUSTER})
  # shellcheck disable=SC2086
  httpbin_pods=$(kubectl get pods -n ${NAMESPACE} -o go-template --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' -l app=httpbin --cluster ${CLUSTER})

  pods=()

  while read -r line; do
    pods+=("$line")
  done <<< "${sleep_pods}"
  while read -r line; do
    pods+=("$line")
  done <<< "${httpbin_pods}"

  if [ ${#pods[@]} = 0 ]; then
    echo "no pods found!"
    exit
  fi

  for pod in "${pods[@]}"
  do
    echo "Dump certs of istio-proxy from pod ${pod}"
    # shellcheck disable=SC2086
    kubectl exec -it -n ${NAMESPACE} "${pod}" -c istio-proxy --cluster ${CLUSTER} -- curl 127.0.0.1:15000/certs
  done
}

case $1 in
  proxyconfig)
    proxyconfig
    ;;

  proxylog)
    proxylog
    ;;

  setproxyloglevel)
    setproxyloglevel
    ;;

  proxycert)
    proxycert
    ;;

  help)
    echo $"Usage: NAMESPACE=<namespace of test workloads> CLUSTER=<test cluster> $0 proxyconfig | proxylog | setproxyloglevel <info|debug|trace> | proxycert"
    ;;

  *)
    echo $"Usage: NAMESPACE=<namespace of test workloads> CLUSTER=<test cluster> $0 proxyconfig | proxylog | setproxyloglevel <info|debug|trace> | proxycert"
esac
