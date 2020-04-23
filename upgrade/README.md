# Istio Upgrade Job

## Upgrade Periodic Prow Job Flow

Currently, we are setting upgrade job prow job trigger in `istio/test-infra` repo.
Under the path of `test-infra/prow/cluster/jobs/all-periodics.yaml`, you will see the defined upgrade job.

Take one defined job for example:

```yaml
# upgrade using istioctl from 1.5 latest to master
- cron: "0 8 * * *"  # starts every day at 08:00AM UTC
  name: istio-upgrade-using-istioctl-1.5_latest-master
  branches: master
  decorate: true
  extra_refs:
    - org: istio
      repo: tools
      base_ref: master
      path_alias: istio.io/tools
  annotations:
    testgrid-dashboards: istio_release-pipeline
    testgrid-alert-email: istio-oncall@googlegroups.com
    testgrid-num-failures-to-alert: '1'
  labels:
    preset-service-account: "true"
  spec:
    containers:
      - <<: *istio_container_with_kind
        env:
          - name: SOURCE_TAG
            value: 1.5_latest
          - name: TARGET_TAG
            value: master
          - name: INSTALL_OPTIONS
            value: istioctl
        command:
          - entrypoint
          - upgrade/run_upgrade_test.sh
    nodeSelector:
      testing: test-pool
```

This job is to test upgrade from Istio 1.5-latest release to Istio master release.
- The `env` section is to pass necessary environment variable to `upgrade/run_upgrade_test.sh`
- The `command` section, we have `entrypoint`, which points to the `istio/tools/upgrade/run_upgrade_test.sh` script.

You can check daily upgrade prow job status from `https://prow.istio.io/?job=istio-upgrade-*`

To add a new prow periodic job, just by following the above example with your passing environment variables and the desired path.

## How to run upgrade test locally

Note: istio/tools has branches `release-1.5` and `master`, there is a slightly difference on Istio-release `tar` name. Since we start to use suffix `linux-amd64.tar.gz` instead of `linux.tar.gz` from a specific point of master (1.6-alpha) release.

Here is to say how to run upgrade test from `1.5-latest` to `master` release.
- Spin up a GKE cluster.
- Go to this link `https://storage.googleapis.com/istio-build/dev/1.5-dev` to get latest `1.5_release_SHA`. 
- Go to this link `https://storage.googleapis.com/istio-build/dev/latest` to get the latest `master_release_SHA`.
- Download corresponding release tar file:
- For 1.5 or before version, use this link:
`
https://storage.googleapis.com/istio-build/dev/
{1.5_release_SHA}/istio-{1.5_release_SHA}-linux.tar.gz`

- For 1.6 or above version, use this link:

`https://storage.googleapis.com/istio-build/dev/
{master_release_SHA}/istio-{master_release_SHA}-linux-amd64.tar.gz`

- export UPGRADE_TEST_LOCAL to any non-empty value to make the test running against your own K8s cluster.





## How to inspect results
