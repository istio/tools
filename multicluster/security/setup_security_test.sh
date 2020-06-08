#!/usr/bin/env bash

# Copyright 2020 Istio Authors

#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at

#       http://www.apache.org/licenses/LICENSE-2.0

#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.

# The values of the following environmental variables should be configured
# based on your multicluster installations (e.g., project, clusters, etc).

# Configure your multicluster project ID, in which you have
# installed Istio multicluster.
export PROJECT_ID=YOUR-PROJECT-ID
# Configure the name of the first cluster in your multicluster project,
# e.g., primary01.
export CLUSTER_1=NAME-OF-YOUR-CLUSTER-1
# Configure the location of the first cluster in your multicluster project,
# e.g., us-central1-a.
export LOCATION_1=LOCATION-OF-YOUR-CLUSTER-1
# Configure the name of the second cluster in your multicluster project,
# e.g., primary02.
export CLUSTER_2=NAME-OF-YOUR-CLUSTER-1
# Configure the location of the second cluster in your multicluster project,
# e.g., us-central1-a.
export LOCATION_2=LOCATION-OF-YOUR-CLUSTER-1

# Configure the download method for Istio release.
# Two methods are supported: curl or gsutil.
export ISTIO_DOWNLOAD_METHOD=curl
# Configure the URL to download Istio release.
# If ISTIO_DOWNLOAD_METHOD=curl, an example value can look like
# https://storage.googleapis.com/gke-release/asm/istio-1.6.4-asm.9-linux-amd64.tar.gz.
# If ISTIO_DOWNLOAD_METHOD=gsutil, an example value can look like
# gs://YOUR-GOOGLE-STORAGE-LINK-FOR-ISTIO-RELEASE.
export ISTIO_RELEASE_URL=https://storage.googleapis.com/gke-release/asm/istio-1.6.4-asm.9-linux-amd64.tar.gz
# Configure the Istio release package name, e.g., istio-1.6.4-asm.9-linux-amd64.tar.gz.
export ISTIO_RELEASE_PKG=istio-1.6.4-asm.9-linux-amd64.tar.gz
# Configure the Istio release name, which should be configured to be
# the same as the directory name after unzipping ISTIO_RELEASE_PKG.
# For example, if unzipping the release pkg istio-1.6.4-asm.9-linux-amd64.tar.gz
# results in the directory istio-1.6.4-asm.9,
# ISTIO_RELEASE_NAME should be configured as istio-1.6.4-asm.9.
export ISTIO_RELEASE_NAME=istio-1.6.4-asm.9
