# How to run the upgrade test

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

You may choose three types of test scenarios (i.e., *upgrade-downgrade*, *upgrade*, and *downgrade*)
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

## Run the upgrade test

    ```bash
    ./run_upgrade_test.sh
    ```
