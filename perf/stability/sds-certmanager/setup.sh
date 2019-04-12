#!/bin/bash

set -ex

WD=$(dirname $0)
WD=$(cd $WD; pwd)

command -v gcloud >/dev/null 2>&1 || {
  echo >&2 "This scenario automatically creates necessary DNS-records in order"
  echo >&2 "to get a obtain a TLS certificate from an ACME-compatible issuer."
  echo >&2 "At the moment the DNS manipulation logic is implemented for Google"
  echo >&2 "Cloud only and requires presence of GCloud SDK and a Google Cloud"
  echo >&2 "project with at least one DNS Zone configured."
  echo >&2
  echo >&2 "Looks like GCloud SDK is not present in the PATH. Aborting..."

  exit 1
}

DNS_ZONE=${DNS_ZONE:?"Name of GCloud DNS zone to use for a new domain record"}

DNS_NAME=$(gcloud dns managed-zones \
  describe $DNS_ZONE --format='value(dnsName)')

if [[ -z "${DNS_NAME}" ]]; then
  echo "Failed to resolve DNS_NAME of the provided DNS_ZONE: ${DNS_ZONE}" 1>&2
  exit 1
else
  INGRESS_DOMAIN="ingress.${NAMESPACE}.ns.${DNS_NAME::-1}"
  echo "The following ingress domain will be configured: ${INGRESS_DOMAIN}"
fi

${WD}/../setup_test.sh "sds-certmanager" "--set namespace=${NAMESPACE:-"sds-certmanager"} --set ingressDomain=${INGRESS_DOMAIN}"

 if [[ -z "${DRY_RUN}" ]]; then
  # Waiting until LoadBalancer is created and retrieving the assigned
  # external IP address.
  echo "Awaiting LoadBalancer creation and fetching assigned external IP..."

  while : ; do
    INGRESS_IP=$(kubectl -n $NAMESPACE \
      get service istio-ingress-$NAMESPACE \
      -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

    if [[ -z "${INGRESS_IP}" ]]; then
      echo "Looks like LoadBalancer creation is still pending, waiting..."
      sleep 5s
    else
      echo "Discovered external IP of the LoadBalancer: ${INGRESS_IP}"
      break
    fi
  done

  echo "Configuring the specified GCloud DNS Zone..."

  if [ -f "${WD}/transaction.yaml" ]; then
    gcloud dns record-sets transaction abort --zone "$DNS_ZONE"
  fi

  gcloud dns record-sets transaction start \
    --zone "$DNS_ZONE"

  OLD_IP=$(gcloud dns record-sets list \
    --zone "$DNS_ZONE" \
    --name "${INGRESS_DOMAIN}." \
    --format "value(rrdatas)")

  if [[ ! -z "${OLD_IP}" ]]; then
    OLD_TTL=$(gcloud dns record-sets list \
      --zone "$DNS_ZONE" \
      --name "${INGRESS_DOMAIN}." \
      --format "value(ttl)")
    OLD_TYPE=$(gcloud dns record-sets list \
      --zone "$DNS_ZONE" \
      --name "${INGRESS_DOMAIN}." \
      --format "value(type)")

    gcloud dns record-sets transaction remove \
      --zone "$DNS_ZONE" \
      --name "${INGRESS_DOMAIN}." \
      --type "${OLD_TYPE}" \
      --ttl "${OLD_TTL}" \
      "${OLD_IP}"
  fi

  gcloud dns record-sets transaction add $INGRESS_IP \
    --zone "$DNS_ZONE" \
    --name "${INGRESS_DOMAIN}." \
    --ttl "60" \
    --type "A"
  gcloud dns record-sets transaction execute \
    --zone $DNS_ZONE
fi
