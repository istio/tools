# Run release qualification test for Anthos Service Mesh

Refer to perf/stability/README.md for more details

run with ASM 1.6 or later on an existing cluster, set INSTALL_ASM=true and these prerequisite: PROJECT_ID, CLUSTER_NAME, CLUSTER_LOCATION

`export RELEASE_URL=https://storage.googleapis.com/gke-release/asm/istio-1.6.5-asm.7-linux-amd64.tar.gz`
`INSTALL_ASM=true RELEASE=release-1.6-asm PROJECT_ID=istio-test CLUSTER_NAME=cluster1 CLUSTER_LOCATION=us-central1-b NAMESPACE_NUM=15 ./long_running_asm.sh --set hub=gcr.io/asm-testing --set tag=1.6.5-asm.7`

run with ASM 1.6 or later with two clusters(for simplicity, we assume two clusters in the same project and same location now)

`export CTX1=cluster1_ctx`
`export CTX2=cluster2_ctx`
`CLUSTER1=cluster1_name`
`CLUSTER1=cluster2_name`
`export RELEASE_URL=https://storage.googleapis.com/gke-release/asm/istio-1.6.5-asm.7-linux-amd64.tar.gz`
`INSTALL_ASM=true MULTI_CLUSTER=true RELEASE=release-1.6-asm PROJECT_ID=istio-test CLUSTER_LOCATION=us-central1-a NAMESPACE_NUM=15 ./long_running_asm.sh --set hub=gcr.io/asm-testing --set tag=1.6.5-asm.7`
