# Istio Multicluster (VPN-mode)

Part of Multicluster stability test scenarios (VPN-mode) that's supposed to be run in the primary stability cluster.

This chart installs simple server applications that aren't aware of the multicluster setup in any way and require no extra configuration. These test apps will be accessed by load generators running in another clusters via a multicluster connectivity, so the access metrics of various scenarios can be collected in the primary stability cluster and compared against each other.

The following scenarios are currently supported:
- `default`: traffic produced by the load generators supposed to be load balanced equally among all available
  endpoints. Example of prometheus metric: `increase(istio_requests_total{destination_app="srv-default"}[1m])/60`. Should constantly produce a value around `200`.
- `locality-distribute` (WIP): traffic produced by the load generators supposed to be load balanced according to the configured weights (25/75). Example of prometheus metric: `increase(istio_requests_total{destination_app="srv-locality-distribute"}[1m])/60`. Should constantly produce a value around `100`.
- `locality-failover` (WIP): traffic produced by the load generators should stay in the external cluster and only spill over in case endpoints availability in the external cluster isn't sufficient. `increase(istio_requests_total{destination_app="srv-locality-failover"}[1m])/60`. Should constantly produce a value that is lower than the corresponding `locality-distribute` value (TBD once the issues are resolved).

See `setup/stability-extra/multicluster-vpn` for the details on how to setup corresponding external clusters with the load generators.
