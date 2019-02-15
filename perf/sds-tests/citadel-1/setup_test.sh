#!/bin/bash

NAMESPACE=${NAMESPACE:?"specify the namespace for running the test"}
NUM=${NUM:?"specify the number of httpbin and sleep workloads"}

function inject_workload() {
  local deployfile="${1:?"please specify the workload deployment file"}"
  # This test uses perf/istio/values-istio-sds-auth.yaml, in which
  # Istio auto sidecar injector is not enabled.
  istioctl kube-inject -f "${deployfile}" -o temp-workload-injected.yaml
  kubectl apply -n ${NAMESPACE} -f temp-workload-injected.yaml
}

TEMP_DEPLOY_NAME="temp_httpbin_sleep_deploy.yaml"
helm template --set replicas="${NUM}" .. > "${TEMP_DEPLOY_NAME}"

inject_workload ${TEMP_DEPLOY_NAME}

echo "Wait 120 seconds for the deployment to be ready ..."
sleep 120


echo "Run curl to test certificate rotations and mTLS"
num_curl=0
num_succeed=0

while [ 1 ]
do
  sleep_pods=$(kubectl get pods -n ${NAMESPACE} -o go-template --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' -l app=sleep)
  pods=()

  while read -r line; do
    pods+=("$line")
  done <<< "${sleep_pods}"

  if [ ${#pods[@]} = 0 ]; then
    echo "no pods found!"
  fi

  for pod in "${pods[@]}"
  do
    resp_code=$(kubectl exec -it -n ${NAMESPACE} "${pod}" -c sleep -- curl -s -o /dev/null -w "%{http_code}" httpbin:8000/headers)
    if [ ${resp_code} = 200 ]; then
      num_succeed=$((num_succeed+1))
    else
      echo "curl from the pod ${pod} failed"
    fi
    num_curl=$((num_curl+1))
    echo "Out of ${num_curl} curl, ${num_succeed} succeeded."
    sleep 1
  done

done
