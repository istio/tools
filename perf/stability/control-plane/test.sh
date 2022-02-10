#!/bin/bash

# Copyright Istio Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

setup() {
  kubectl create ns twopods-istio
  kubectl create ns invalid-namespace-0
  kubectl apply -f <(istioctl kube-inject -f ./samples/sleep.yaml) -n invalid-namespace-0
  kubectl apply -f <(istioctl kube-inject -f ./samples/httpbin.yaml) -n twopods-istio
  go build ../../benchmark/security/generate_policies/generate_policies.go ../../benchmark/security/generate_policies/generate.go ../../benchmark/security/generate_policies/jwt.go
}

checkResourcesInCluster() {
  end=$((SECONDS+30))
  failed=true
  while [ $SECONDS -lt $end ]; do
      local status
      status=$(kubectl exec "$(kubectl get pod -l app=sleep -n invalid-namespace-0 -o jsonpath={.items..metadata.name})" -c sleep -n invalid-namespace-0 -- curl http://httpbin.twopods-istio:8000/ip -sS -o /dev/null -w "%{http_code}\n")
      if [ "$status" -eq 403 ]
      then
        failed=false
        break
      fi
      :
  done
  
  if [ "$failed" = true ]
  then
    echo "deny policies are not deployed properly"
    exit 1
  fi
}

addPolicy() {
  ./generate_policies -configFile="config.json" > deployment.yaml
  kubectl apply -f deployment.yaml
  checkResourcesInCluster
  
  
  ./generate_policies -configFile="config_100.json" > deployment_100.yaml
  kubectl apply -f deployment_100.yaml
  checkResourcesInCluster
  

  ./generate_policies -configFile="config_1000.json" > deployment_1000.yaml
  kubectl apply -f deployment_1000.yaml
  checkResourcesInCluster
}

updatePolicy() {
  kubectl delete AuthorizationPolicy --all -n twopods-istio

  # update the number of namespaces on which authz policy applies
  ./generate_policies -configFile="config_update_ns.json" > deployment_update_ns.yaml
  checkResourcesInCluster

  kubectl delete AuthorizationPolicy --all -n twopods-istio

  # update the number of paths on which authz policy applies
  ./generate_policies -configFile="config_update_paths.json" > deployment_update_paths.yaml
  checkResourcesInCluster
}

cleanCluster() {
  kubectl delete ns twopods-istio
  kubectl delete ns invalid-namespace-0
}

for iter in {1..6}
do
  echo "-------- Iteration ${iter} starts ----------"
  setup
  addPolicy
  updatePolicy
  cleanCluster
  echo "-------- Iteration ${iter} ends ----------"
done




