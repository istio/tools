# SDS + CertManager

This scenario verifies the use of Istio IngressGateway as standard K8S Ingress
backend with SDS enabled and real TLS certificates managed by the CertManager.
CertManager uses 'http01' mode and configures an Ingress (backed by Istio) in
order to pass ACME challenge.

This configuration requires a real GCloud DNS Zone setup in the current GCloud
project. The setup script will use specified zone to create a subdomain based
on the specified namespace pointing to the external IP of the Ingress Gateway.
A TLS certificate will be issued for the subdomain and a Fortio Client will
simulate external HTTPS traffic hitting internal Fortio Server to verify the
connectivity.

Use the following command to (re)install the scenario into a new/existing
namespace (set `DRY_RUN=1` for a dry run):

```bash
NAMESPACE=sds-certmanager DNS_ZONE=myzone ./setup.sh
```

Assuming "myzone" has an "example.com" domain associated with it, the following
subdomain will be configured and used in this scenario:

```plain
ingress.sds-certmanager.ns.example.com
```

Certificates are periodically rotated to verify that SDS is able to detect
the renewed certificate and hot-swap the cert used by Ingress Gateway. The
Fortio client is configured to abort on connectivity errors, so in order to
verify the behavior check that the certificate was issued within last 30 min
and 'fortio-client' pod hasn't been restarted (i.e. ensure that pod restarts
do not correlate with certificate rotations).
