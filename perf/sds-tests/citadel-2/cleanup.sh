#!/bin/bash

kubectl delete deploy httpbin sleep
kubectl delete svc httpbin sleep
kubectl delete serviceaccount sleep

# If you need to delete the Istio deployment, run the following command also.
# kubectl delete ns istio-system
