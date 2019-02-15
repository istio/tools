#!/bin/bash	
WD=$(dirname $0)	
WD=$(cd "${WD}"; pwd)	
cd "${WD}"	

 set -ex	

 NAMESPACE=${1:?"namespace"}	
NAMEPREFIX=${2:?"prefix name for service. typically svc-"}	

 # Get pod ip range, there must be a better way, but this works.	
function ip_range() {	
    kubectl get pods --namespace kube-system -o wide | grep kube-dns | awk '{print $6}'|head -1 | awk -F '.' '{printf "%s.%s.0.0/16\n", $1, $2}'	
}	

 function run_test() {	
  YAML=$(mktemp).yml	
  helm -n ${NAMESPACE} template \	
	  --set serviceNamePrefix="${NAMEPREFIX}" \	
    --set Namespace="${NAMESPACE}" \	
          . > "${YAML}"	
  echo "Wrote ${YAML}"	

   kubectl create ns "${NAMESPACE}" || true	
  kubectl label namespace "${NAMESPACE}" istio-injection=enabled --overwrite	

   # remove stdio rules	
  kubectl --namespace istio-system delete rules stdio stdiotcp || true	

   if [[ -z "${DELETE}" ]];then	
    sleep 3	
    kubectl -n "${NAMESPACE}" apply -f "${YAML}"	
  else	
    kubectl -n "${NAMESPACE}" delete -f "${YAML}" || true	
    kubectl delete ns "${NAMESPACE}"	
  fi	
}	

 run_test
