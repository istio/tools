The Lua filter is injected before the Istio `jwt-auth` filter. If a JWT token is presented on an HTTP request, the Lua filter will check if the JWT token header contains `alg:ES256`, and if so, reject the request.

To install the Lua filter, please invoke the following commands:

```bash
$ git clone git@github.com:istio/tools.git
$ cd tools/examples/luacheck/
$ ./setup.sh 
```
