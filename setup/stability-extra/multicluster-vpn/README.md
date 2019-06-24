# Istio Multicluster (VPN-mode)

A set of external cluster install configurations for Multicluster (VPN-mode) test scenarios.

For each template file under `perf/stability/multicluster-vpn/templates/` there is a corresponding directory under
`setup/stability-extra/multicluster-vpn/` containing a set of test apps and load generators along with installation script that spins up an isolated Istio control plane with multicluster configuration pointing to the primary stability cluster. This allows to test various multicluster configuration options within the same cluster.

In order to install all scenarios into an external cluster and configure multicluster connectivity with the primary stability cluster, use the `setup.sh` script in the current directory. Point script to the primary stability cluster (as the first cluster) and to the newly provisioned external cluster (as the second cluster). Make sure that the two clusters reside in different regions so that locality-aware load balancing can be exercised.

See the `setup.sh` script (header) for the arguments reference and usage example.
