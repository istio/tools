Multi Cluster VPC Suite tests Istio Multi Cluster feature with shared VPC.


To setup the environment and deploy BookInfo app to two cluters.

```
RELEASE="release-1.1-20190209-09-16" proj="your-gcp-project" \
zone="us-central1-a" cluster1="cluster1" cluster2="cluster2" \
  ./setup.sh setup
```

This will create two GKE clusters with IP aliasing enabled and install Istio
accordingly in each cluster. And finally deploy BookInfo app in two clusters.

To tear down the clusters and clean up the resources

```
RELEASE="release-1.1-20190209-09-16" proj="your-gcp-project" \
zone="us-central1-a" cluster1="cluster1" cluster2="cluster2" \
  ./setup.sh cleanup
```
