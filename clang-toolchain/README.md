# clang-toolchain

Because [llvm+clang](https://github.com/llvm/llvm-project/releases) doesn't ship binaries for CentOS 7,
we need to build them ourselves. Previously these artifacts were published by [getenvoy-package](https://github.com/tetratelabs-attic/getenvoy-package/)
(which is where these scripts are from).

To run this build in your own cloudbuild environment:

1. Fork this repository, you'll need to be the repo owner to set up cloud build triggers.
1. Create a [GCP project](https://cloud.google.com/) if you don't already have one. You'll also need to enable
   [Cloud Build](https://cloud.google.com/build) and [Cloud Storage](https://cloud.google.com/storage)
1. Create a Cloud Storage bucket to store the resulting artifacts.
1. Modify `cloudbuild.yaml` in your fork to point to your bucket.
1. Create a Cloud Build trigger pointing to your fork. The path to the configuration file is
   `clang-toolchain/cloudbuild.yaml`. The trigger type can be "Manual Trigger".
1. Trigger a build and after about 2 hours, you should have the artifacts pushed to your GCS bucket.
