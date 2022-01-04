#!/bin/bash

checkResourcesInCluster() {
    kubectl label namespace invalid-namespace-0 istio-injection=enabled --overwrite
    kubectl apply -f <(istioctl kube-inject -f ../../istio-install/tmp/istio-1.12.0/samples/sleep/sleep.yaml) -n invalid-namespace-0
    kubectl apply -f <(istioctl kube-inject -f ../../istio-install/tmp/istio-1.12.0/samples/httpbin/httpbin.yaml) -n twopods-istio
    kubectl exec "$(kubectl get pod -l app=sleep -n invalid-namespace-0 -o jsonpath={.items..metadata.name})" -c sleep -n invalid-namespace-0 -- curl http://httpbin.twopods-istio:8000/ip -sS -o /dev/null -w "%{http_code}\n"
}

getMetrics() {
    echo "inside"
    end=$((SECONDS+$1))
    while [ $SECONDS -lt $end ]; 
    do
      ./promql-cli/promql --host "http://localhost:9090" 'sum(container_memory_working_set_bytes{namespace="istio-system", container="discovery"}) BY (pod)' >> memory_usage.txt
      ./promql-cli/promql --host "http://localhost:9090" 'sum(rate(container_cpu_usage_seconds_total{namespace="istio-system", container="discovery"}[5m])) BY (pod)' >> cpu_load.txt
      :
    done
}

addPolicy() {
  kubectl create ns twopods-istio
  kubectl create ns invalid-namespace-0
  go run ../../benchmark/security/generate_policies/generate_policies.go ../../benchmark/security/generate_policies/generate.go ../../benchmark/security/generate_policies/jwt.go -configFile="config.json" > deployment.yaml
  kubectl apply -f deployment.yaml
  checkResourcesInCluster
  getMetrics 30
  
  go run ../../benchmark/security/generate_policies/generate_policies.go ../../benchmark/security/generate_policies/generate.go ../../benchmark/security/generate_policies/jwt.go -configFile="config_100.json" > deployment_100.yaml
  kubectl apply -f deployment_100.yaml
  checkResourcesInCluster
  getMetrics 300

  go run ../../benchmark/security/generate_policies/generate_policies.go ../../benchmark/security/generate_policies/generate.go ../../benchmark/security/generate_policies/jwt.go -configFile="config_1000.json" > deployment_1000.yaml
  kubectl apply -f deployment_1000.yaml
  checkResourcesInCluster
  getMetrics 600
}

updatePolicy() {
  kubectl delete AuthorizationPolicy --all -n twopods-istio

  # update the number of namespaces on which authz policy applies
  go run ../../benchmark/security/generate_policies/generate_policies.go ../../benchmark/security/generate_policies/generate.go ../../benchmark/security/generate_policies/jwt.go -configFile="config_update_ns.json" > deployment_update_ns.yaml
  checkResourcesInCluster
  getMetrics 600

  kubectl delete AuthorizationPolicy --all -n twopods-istio

  # update the number of paths on which authz policy applies
  go run ../../benchmark/security/generate_policies/generate_policies.go ../../benchmark/security/generate_policies/generate.go ../../benchmark/security/generate_policies/jwt.go -configFile="config_update_paths.json" > deployment_update_paths.yaml
  checkResourcesInCluster
  getMetrics 600
}


cleanCluster() {
  kubectl delete ns twopods-istio
  kubectl delete ns invalid-namespace-0
  getMetrics 100
}

addPolicy
updatePolicy
cleanCluster




