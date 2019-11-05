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
# It offers the following options:
# - check: Check the current root certificate lifetime.
# - transition: Extend the lifetime of the current root certificate.
# - verify: Check the new workload certificates are generated.
# This script requires openssl, kubectl and bc.

trustdomain() {
  # shellcheck disable=SC2086
  openssl x509 -in ${1} -noout -issuer | cut -f3 -d'='
}

check_secret () {
	# shellcheck disable=SC2006
	# shellcheck disable=SC2086
	MD5=`kubectl get secret $1 -o yaml -n $2 | sed -n 's/^.*root-cert.pem: //p' | md5sum | awk '{print $1}'`
	if [ "$ROOT_CERT_MD5" != "$MD5" ]; then
		echo "  Secret $2.$1 is DOES NOT match current root."
		NOT_UPDATED="$NOT_UPDATED $2.$1"
	else
		echo "  Secret $2.$1 matches current root."
	fi
}

verify_namespace () {
	# shellcheck disable=SC2006
	# shellcheck disable=SC2086
	SECRETS=`kubectl get secret -n $1 | grep "istio\.io\/key-and-cert" | awk '{print $1}'`
	for s in $SECRETS
	do
		# shellcheck disable=SC2086
		check_secret $s $1
	done
}

verify() {
  NOT_UPDATED=

  echo "This script checks the current root CA certificate is propagated to all the Istio-managed workload secrets in the cluster."

  # shellcheck disable=SC2006
  ROOT_SECRET=`kubectl get secret istio-ca-secret -o yaml -n istio-system | sed -n 's/^.*ca-cert.pem: //p'`
  # shellcheck disable=SC2086
  if [ -z $ROOT_SECRET ]; then
    echo "Root secret is empty. Are you using the self-signed CA?"
    exit
  fi

  # shellcheck disable=SC2006
  ROOT_CERT_MD5=`kubectl get secret istio-ca-secret -o yaml -n istio-system | sed -n 's/^.*ca-cert.pem: //p' | md5sum | awk '{print $1}'`

  # shellcheck disable=SC2086
  echo Root cert MD5 is $ROOT_CERT_MD5

  # shellcheck disable=SC2006
  NS=`kubectl get ns | grep -v "STATUS" | grep -v "kube-system" | grep -v "kube-public" | awk '{print $1}'`

  for n in $NS
  do
    echo "Checking namespace: $n"
    # shellcheck disable=SC2086
    verify_namespace $n
  done

  if [ -z "$NOT_UPDATED" ]; then
    echo "=====All Istio mutual TLS keys and certificates match the current root!====="
    echo
  else
    echo "=====The following secrets do not match current root: ====="
    echo $NOT_UPDATED
    echo
  fi
}

check() {
  # shellcheck disable=SC2006
  ROOT_SECRET=`kubectl get secret istio-ca-secret -o yaml -n istio-system | sed -n 's/^.*ca-cert.pem: //p'`
  # shellcheck disable=SC2086
  if [ -z $ROOT_SECRET ]; then
    echo "Root secret is empty. Are you using the self-signed CA?"
    return
  fi

  echo "Fetching root cert from istio-system namespace..."
  kubectl get secret -n istio-system istio-ca-secret -o yaml | awk '/ca-cert/ {print $2}' | base64 --decode > ca.cert
  if [[ ! -f ./ca.cert ]]; then
    echo "failed to get cacert, check the istio installation namespace."
    return
  fi

  rootDate=$(openssl x509 -in ca.cert -noout -enddate | cut -f2 -d'=')
  if [[ "$(uname)" == "Darwin" ]]; then
    rootSec=$(date -jf "%b  %e %k:%M:%S %Y %Z" "${rootDate}" '+%s')
  else
    rootSec=$(date -d "${rootDate}" '+%s')
  fi

  # shellcheck disable=SC2006
  nowSec=`date '+%s'`
  remainDays=$(echo "(${rootSec} - ${nowSec}) / (3600 * 24)" | bc)

  cat << EOF
Your Root Cert will expire after
   ${rootDate}
Current time is
  $(date)


=====YOU HAVE ${remainDays} DAYS BEFORE THE ROOT CERT EXPIRES!=====

EOF
}

transition() {
  # Get root cert and private key and generate a 10 year root cert:
  kubectl get secret istio-ca-secret -n istio-system -o yaml | sed -n 's/^.*ca-cert.pem: //p' | base64 --decode > old-ca-cert.pem
  kubectl get secret istio-ca-secret -n istio-system -o yaml | sed -n 's/^.*ca-key.pem: //p' | base64 --decode > ca-key.pem

  TRUST_DOMAIN="$(echo -e "$(trustdomain old-ca-cert.pem)" | sed -e 's/^[[:space:]]*//')"
  echo "Create new ca cert, with trust domain as $TRUST_DOMAIN"
  openssl req -x509 -new -nodes -key ca-key.pem -sha256 -days 3650 -out new-ca-cert.pem -subj "/O=${TRUST_DOMAIN}"

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
  OUT=`kubectl get po -n $1 -o yaml | grep "istio\/proxyv2\:1\."`

  for LINE in $OUT;
  do
    if [[ $LINE == *"istio/proxyv2"* ]]; then
      LINE=${LINE#"gke.gcr.io/istio/proxyv2:"};
      LINE=${LINE#"docker.io/istio/proxyv2:"};
      LINE=${LINE#"istio/proxyv2:"};
      VER=${LINE%%"-gke.0"};
      echo "Istio proxy version: $VER";
    fi
  done
}

check_version() {
  # shellcheck disable=SC2006
  NS=`kubectl get ns | grep -v "STATUS" | grep -v "kube-system" | grep -v "kube-public" | awk '{print $1}'`

  for n in $NS
  do
    echo "Checking namespace: $n"
    # shellcheck disable=SC2086
    check_version_namespace $n
  done
}

case $1 in
  check-root)
    check
    ;;

  check-version)
    check_version
    ;;

  root-transition)
    transition
    ;;

  verify-certs)
    verify
    ;;

  *)
    echo $"Usage: check-root | check-version | root-transition | verify-certs

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
