# HTTP 1.0 Testing

This test the Envoy proxies will accept.

This is tested by sending http 1.0 requests repeatedly.

HTTP 1.0 needs to be enabled in pilot for this to work. This can be done by setting `pilot.env.PILOT_HTTP10=1`.