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

# This script extends the lifetime of the self-signed Citadel root certificate in the current cluster.
# This script requires openssl, kubectl and bc.

trustdomain() {
  openssl x509 -in "$1" -noout -issuer | cut -f3 -d'='
}

check_secret () {
  local md5
  md5=$(kubectl get secret "$1" -o yaml -n "$2" | sed -n \
    's/^.*root-cert.pem: //p' | md5sum | awk '{print $1}')
	if [ "${ROOT_CERT_MD5}" != "${md5}" ]; then
		echo "  Secret $2.$1 is DOES NOT match current root."
		NOT_UPDATED="${NOT_UPDATED} $2.$1"
	else
		echo "  Secret $2.$1 matches current root."
	fi
}

verify_namespace () {
  local secrets
  secrets=$(kubectl get secret -n "$1" | grep "istio\.io\/key-and-cert" | awk '{print $1}')
	for s in ${secrets}
	do
		check_secret "$s" "$1"
	done
}

verify_certs() {
  NOT_UPDATED=

  echo "This script checks the current root CA certificate is propagated to all the Istio-managed workload secrets in the cluster."

  local root_secret
  root_secret=$(kubectl get secret istio-ca-secret -o yaml -n istio-system \
  | sed -n 's/^.*ca-cert.pem: //p')
  if [ -z "${root_secret}" ]; then
    echo "Root secret is empty. Are you using the self-signed CA?"
    exit
  fi

  ROOT_CERT_MD5=$(kubectl get secret istio-ca-secret -o yaml -n istio-system \
  | sed -n 's/^.*ca-cert.pem: //p' | md5sum | awk '{print $1}')

  echo "Root cert MD5 is ${ROOT_CERT_MD5}"

  local ns
  ns=$(kubectl get ns | grep -v "STATUS" | grep -v "kube-system" | grep -v "kube-public" | awk '{print $1}')

  for n in ${ns}
  do
    echo "Checking namespace: ${n}"
    verify_namespace "${n}"
  done

  if [ -z "${NOT_UPDATED}" ]; then
    echo
    echo "=====All Istio mutual TLS keys and certificates match the current root!====="
    echo
  else
    echo
    echo "=====The following secrets do not match current root: ====="
    echo ${NOT_UPDATED}
    echo
  fi
}

check_root() {
  local root_secret
  root_secret=$(kubectl get secret istio-ca-secret -o yaml -n istio-system \
  | sed -n 's/^.*ca-cert.pem: //p')
  if [ -z "${root_secret}" ]; then
    echo "Root secret is empty. Are you using the self-signed CA?"
    return
  fi

  echo "Fetching root cert from istio-system namespace..."
  kubectl get secret -n istio-system istio-ca-secret -o yaml | \
    awk '/ca-cert/ {print $2}' | base64 --decode > ca.cert
  if [[ ! -f ./ca.cert ]]; then
    echo "failed to get cacert, check the istio installation namespace."
    return
  fi

  local root_date
  local root_sec
  root_date=$(openssl x509 -in ca.cert -noout -enddate | cut -f2 -d'=')
  if [[ "$(uname)" == "Darwin" ]]; then
    root_sec=$(date -jf "%b  %e %k:%M:%S %Y %Z" "${root_date}" '+%s')
  else
    root_sec=$(date -d "${root_date}" '+%s')
  fi

  local now_sec
  local days_left
  now_sec=$(date '+%s')
  days_left=$(echo "(${root_sec} - ${now_sec}) / (3600 * 24)" | bc)

  cat << EOF
Your Root Cert will expire after
   ${root_date}
Current time is
  $(date)


=====YOU HAVE ${days_left} DAYS BEFORE THE ROOT CERT EXPIRES!=====

EOF
}

root_transition() {
  # Get root cert and private key and generate a 10 year root cert:
  kubectl get secret istio-ca-secret -n istio-system -o yaml | sed -n 's/^.*ca-cert.pem: //p' | base64 --decode > old-ca-cert.pem
  kubectl get secret istio-ca-secret -n istio-system -o yaml | sed -n 's/^.*ca-key.pem: //p' | base64 --decode > ca-key.pem

  local trust_domain
  trust_domain="$(echo -e "$(trustdomain old-ca-cert.pem)" | sed -e 's/^[[:space:]]*//')"
  echo "Create new ca cert, with trust domain as ${trust_domain}"
  openssl req -x509 -new -nodes -key ca-key.pem -sha256 -days 3650 -out new-ca-cert.pem -subj "/O=${trust_domain}"

  echo "$(date) delete old CA secret"
  kubectl -n istio-system delete secret istio-ca-secret
  echo "$(date) create new CA secret"
  kubectl create -n istio-system secret generic istio-ca-secret --from-file=ca-key.pem=ca-key.pem --from-file=ca-cert.pem=new-ca-cert.pem --type=istio.io/ca-root

  echo "$(date) Restarting Citadel ..."
  kubectl delete pod -l istio=citadel -n istio-system

  echo "$(date) restarted Citadel, checking status"
  kubectl get pods -l istio=citadel -n istio-system

  echo "New root certificate:"
  openssl x509 -in new-ca-cert.pem -noout -text

  echo "Your old certificate is stored as old-ca-cert.pem, and your private key is stored as ca-key.pem"
  echo "Please save them safely and privately."
}

check_version_namespace() {
  local out
  local line
  local ver
  out=$(kubectl get po -n "$1" -o yaml | grep "proxyv2\:1\.")

  for line in ${out};
  do
    if [[ ${line} == *"proxyv2"* ]]; then
      line=${line#"gke.gcr.io/istio/proxyv2:"};
      line=${line#"docker.io/istio/proxyv2:"};
      line=${line#"gcr.io/gke-release/istio/proxyv2:"};
      line=${line#"istio/proxyv2:"};
      line=${line#"gcr.io/istio-testing/proxyv2:"};
      ver=${line%%"-gke.0"};
      echo "Istio proxy version: $ver";
    fi
  done
}

check_version() {
  local ns
  ns=$(kubectl get ns | grep -v "STATUS" | grep -v "kube-system" | grep -v \
    "kube-public" | awk '{print $1}')

  for n in ${ns}
  do
    echo "Checking namespace: ${n}"
    check_version_namespace "${n}"
  done
}

case $1 in
  check-root)
    check_root
    ;;

  check-version)
    check_version
    ;;

  root-transition)
    root_transition
    ;;

  verify-certs)
    verify_certs
    ;;

  *)
    echo "Usage: check-root | check-version | root-transition | verify-certs

check-root
  Check the expiration date of the current root certificate.

check-version
  Check the version of all Istio sidecars in the system.

root-transition
  Conduct a root cert transition. This will replace the current root
  certificate with a new 10-year lifetime root certificate. Use caution when
  running this command since it modifies your cluster.

verify-certs
  Verify that the current root certificate is propagated to every workload's
  secret.
"

esac
