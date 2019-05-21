This directory uses kustomize to deploy a weekly job to pull the latest daily build of a given version of istio to one of our release qualification clusters.  The job runs every Friday night at 8PM Pacific.  To use kustomize embedded in kubectl, you will need kubeclt v 1.14+.  If you are unable to run on a recent kubectl, you can install kustomize separately and replace the `kubectl apply -k <folder>` commands below with `kustomize build <folder> | kubectl apply -f -`.

The base kustomization deploys a job targeting the latest daily v1.1 build of Istio.  To deploy this job, run `kubectl apply -k base`.

To target version 1.2 of istio, run `kubectl apply -k overlays/v1.2`

Other target versions can be added using new overlay folders.

To run the generated job immediately, rather than waiting for the next scheduled run, use `kubectl create job --from=cronjob/qual-test-update qual-test-manual`, and track progress with `kubectl logs $(kubectl get po -l app=qual-test-update -o jsonpath='{.items[0].metadata.name}') -f`