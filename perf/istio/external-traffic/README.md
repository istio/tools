# External Traffic

This tests external traffic without using `ServiceEntry`s.

Instead, `global.outboundTrafficPolicy.mode=ALLOW_ANY` is set and a Sidecar resource is created.

## Setup

Helm must be installed with `--set global.outboundTrafficPolicy.mode=ALLOW_ANY` or this will not work.

Once this is done, run `./setup.sh`
