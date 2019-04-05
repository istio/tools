#!/bin/bash

#!/bin/bash

set -xe
NAMESPACE=${NAMESPACE:?"specify the namespace to delete"}

kubectl delete ns ${NAMESPACE}

# If you need to delete the Istio deployment, run the following command also.
# kubectl delete ns istio-system
