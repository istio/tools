# Wasm plugin testing

In this test scenario,
- A client sends requests repeatedly with 100ms delay between the requests, and the server makes echo responses.
- WasmPlugin is also repeatedly changed for each 10 seconds.
- If Istio can interpret the proxy metadata "WASM_PURGE_INTERVAL" and "WASM_MODULE_EXPIRY", the Wasm binaries are expired within 5 seconds. So, for each request, Wasm binaries are pulled.

With such frequent changes of WasmPlugin, the traffic between the echo client/server should be transferred without any errors or degrading the performance.
