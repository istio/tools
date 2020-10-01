# How to run the upgrade and downgrade test

## Create a test cluster

Create a cluster and set it as the current cluster.
You may use `kubectl config current-context` to confirm that the cluster
is as expected.

## Configure the test options

Let the root directory of this repo be *ROOT-OF-REPO*.
Enter the directory containing the tests.

    ```bash
    cd ROOT-OF-REPO/upgrade
    ```

You may choose three types of test scenarios for in-place upgrades (i.e., *upgrade-downgrade*, *upgrade*, and *downgrade*)
and two types of test scenarios for dual control-plane upgrades (i.e., *dual-control-plane-upgrade*, *dual-control-plane-rollback*)
by configuring the TEST_SCENARIO variable. In the following example commands,
*SOURCE_TAG* specifies the
version of Istio to be upgraded/downgraded and *TARGET_TAG* specifies the
version of Istio the test will be upgraded/downgraded to.

Note that you may need to configure the variables in the commands based
on your test cases.

* When TEST_SCENARIO is configured as *upgrade-downgrade*,
Istio will be upgraded and then downgraded.
This is the default test flow. The following is an example command
to configure this test scenario:

    ```bash
    export TEST_SCENARIO=upgrade-downgrade; export SOURCE_TAG=1.5_latest; export TARGET_TAG=master; export INSTALL_OPTIONS=istioctl; export UPGRADE_TEST_LOCAL=true;
    ```

* When TEST_SCENARIO is configured as *upgrade*,
Istio will be upgraded. The following is an example command
to configure this test scenario:

    ```bash
    export TEST_SCENARIO=upgrade; export SOURCE_TAG=1.5_latest; export TARGET_TAG=master; export INSTALL_OPTIONS=istioctl; export UPGRADE_TEST_LOCAL=true;
    ```

* When TEST_SCENARIO is configured as *downgrade*,
Istio will be downgraded. The following is an example command
to configure this test scenario:

    ```bash
    export TEST_SCENARIO=downgrade; export SOURCE_TAG=master; export TARGET_TAG=1.5_latest; export INSTALL_OPTIONS=istioctl; export UPGRADE_TEST_LOCAL=true;
    ```

* When TEST_SCENARIO is configured as *dual-control-plane-upgrade*,
Istio will first install control plane specified by SOURCE_TAG, then
install control plane setting revision to TARGET_TAG. It will restart
deployments one by one to point to the control plane with TARGET_TAG
before uninstalling the one with SOURCE_TAG

    ```bash
    export TEST_SCENARIO=dual-control-plane-upgrade
    export SOURCE_TAG=1.7_latest
    export TARGET_TAG=master
    export UPGRADE_TEST_LOCAL=true
    ```

* When TEST_SCENARIO is configured as *dual-control-plane-rollback*
Istio will first install control plane of version SOURCE_TAG and then
another one with TARGET_TAG. Both of then will be running simultaneously.
It then upgrades one of the deployments to point to TARGET_TAG control plane.
Then the deployment is again restarted to point to SOURCE_TAG and finally,
the control plane running version TARGET_TAG is uninstalled

    ```bash
    export TEST_SCENARIO=dual-control-plane-rollback
    export SOURCE_TAG=1.7_latest
    export TARGET_TAG=master
    export UPGRADE_TEST_LOCAL=true
    ```

## Run the upgrade or downgrade test

    ```bash
    ./run_upgrade_downgrade_test.sh
    ```
