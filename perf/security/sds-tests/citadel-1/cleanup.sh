#!/bin/bash

#!/bin/bash

set -xe
NAMESPACE=${NAMESPACE:?"specify the namespace to delete"}
CLUSTER=${CLUSTER:?"specify the cluster for running the test"}

kubectl delete ns ${NAMESPACE} --cluster ${CLUSTER}

# If you need to delete the Istio deployment, run the following command also.
# kubectl delete ns istio-system --cluster ${CLUSTER}
