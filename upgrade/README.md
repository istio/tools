# How to run the upgrade test

- Create a cluster and set it as the current cluster.
You may use `kubectl config current-context` to confirm that the cluster
is as expected.

- Configure the upgrade test.
Let the root directory of this repo be *ROOT-OF-REPO*.
Run the following commands, in which *SOURCE_TAG* specifies the
version of Istio to be upgraded and *TARGET_TAG* specifies the
version of Istio to which the test will upgrade.
Note that you may need to configure the variables in the command based
on your test cases.

    ```bash
    cd ROOT-OF-REPO/upgrade
    export SOURCE_TAG=1.5_latest; export TARGET_TAG=master; export INSTALL_OPTIONS=istioctl; export UPGRADE_TEST_LOCAL=true;
    ```

- Run the upgrade test.

    ```bash
    ./run_upgrade_test.sh
    ```
