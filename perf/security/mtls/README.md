# Mutual TLS Performance Evaluation

This test suite measures the performance overhead due to the mutual TLS encryption, specifically
the data plane CPU, latency and throughput.

**Test Setup**

- Istio is installed with auto mTLS enabled. DestinationRule ISTIO_MUTUAL is to be deprecated for mTLS.
Relying on auto mTLS is more future proof.
- Deployment is described via service graph.
- Load client is sending traffic to two group of service in the service graph:
  1. mtls frontend -> mtls backend.
  1. plaintext frontend -> plaintext backend


Notes

- `curl http://35.224.165.189/ -H 'Host: svc-0.local'  -v`, 503, due to svc-01 is unhealthy. readiness probe fails.
   1. reason for the readiness fail, wasm plugin fails with log

   ```
   0.0.0.0_9091: Proto constraint validation failed (WasmValidationError.Config: ["embedded message failed validation"] | caused by PluginConfigValidationError.VmConfig: ["embedded message failed validation"] | caused by VmConfigValidationError.Code: ["embedded message failed validation"] | caused by field: "specifier", reason: is required): config {
    vm_config {
      runtime: "envoy.wasm.runtime.null"
      code {
      }
    }
    configuration: "envoy.wasm.metadata_exchange"
  }
  , 10.60.11.207_15029: Proto constraint validation failed (WasmValidationError.Config: ["embedded message failed validation"] | caused by PluginConfigValidationError.VmConfig: ["embedded message failed validation"] | caused by VmConfigValidationError.Code: ["embedded message failed validation"] | caused by field: "specifier", reason: is required): config {
    vm_config {
      runtime: "envoy.wasm.runtime.null"
      code {
      }
    }
    configuration: "envoy.wasm.metadata_exchange"
  }
  , virtualInbound: Proto constraint validation failed (WasmValidationError.Config: ["embedded message failed validation"] | caused by PluginConfigValidationError.VmConfig: ["embedded message failed validation"] | caused by VmConfigValidationError.Code: ["embedded message failed validation"] | caused by field: "specifier", reason: is required): config {
    vm_config {
      runtime: "envoy.wasm.runtime.null"
      code {
      }
    }
    configuration: "envoy.wasm.metadata_exchange"
  }
   ```
- client fails with 404 error? gw is the same as automtls.