#!/bin/bash
kubectl apply -n twopods-istio -f https://raw.githubusercontent.com/istio/istio/release-1.15/samples/extauthz/ext-authz.yaml
kubectl patch configmap -n istio-system istio --patch-file ext-authz_patch.yaml
kubectl rollout restart deployment/istiod -n istio-system
