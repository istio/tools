#!/bin/bash -e

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

# Default wait timeout is 180 seconds
set +x
count=0

#!/bin/bash
function wait_for_control_plane() {
   pods=("$@")

   for x in "${pods[@]}"; do
       echo "$x"
       pod_name=""
       while [ "x$pod_name" == "x" ]; do
            pod_name=$(kubectl get pod -n kube-system | grep -G ^"$x" | awk '{print $1}')
            if [ "x$pod_name" != "x" ]; then
                while true; do
                    pod_status=$(kubectl get pod -n kube-system "$pod_name" -o jsonpath='{.status.phase}')
                    echo "Pod name: $pod_name status: $pod_status"
                    if [ "x$pod_status" != "xRunning" ]; then
                        if [ $count -gt 180 ]; then
                           echo "Kubernetes cluster control plane failed to come up"
                           return 1
                        fi
                        count=$((count+1))
                        sleep 1
                    else
                        break
                    fi
                done
            else
                if [ $count -gt 180 ]; then
                    echo "Kubernetes cluster control plane failed to come up"
                    return 1
                fi
                count=$((count+1))
                sleep 1
            fi
        done
    done
    return 0
}