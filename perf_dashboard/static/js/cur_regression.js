window.onload = function () {
  // p90
  let latency_none_mtls_both_p90_trending = constructDataSeries(latency_none_mtls_both_p90, latency_none_mtls_base_p90)
  let latency_none_plaintext_both_p90_trending = constructDataSeries(latency_none_plaintext_both_p90, latency_none_mtls_base_p90)
  let latency_v2_stats_nullvm_both_p90_trending = constructDataSeries(latency_v2_stats_nullvm_both_p90, latency_none_mtls_base_p90)
  let latency_v2_stats_wasm_both_p90_trending = constructDataSeries(latency_v2_stats_wasm_both_p90, latency_none_mtls_base_p90)
  let latency_v2_sd_nologging_nullvm_both_p90_trending = constructDataSeries(latency_v2_sd_nologging_nullvm_both_p90, latency_none_mtls_base_p90)
  let latency_v2_sd_full_nullvm_both_p90_trending = constructDataSeries(latency_v2_sd_full_nullvm_both_p90, latency_none_mtls_base_p90)

  p90ModesData = [];
  p90ModesData.push(latency_none_mtls_both_p90_trending);
  p90ModesData.push(latency_none_plaintext_both_p90_trending);
  p90ModesData.push(latency_v2_stats_nullvm_both_p90_trending);
  p90ModesData.push(latency_v2_stats_wasm_both_p90_trending);
  p90ModesData.push(latency_v2_sd_nologging_nullvm_both_p90_trending);
  p90ModesData.push(latency_v2_sd_full_nullvm_both_p90_trending);

  generateChart("trending_p90", p90ModesData);

  // p99
  let latency_none_mtls_both_p99_trending = constructDataSeries(latency_none_mtls_both_p99, latency_none_mtls_base_p99)
  let latency_none_plaintext_both_p99_trending = constructDataSeries(latency_none_plaintext_both_p99, latency_none_mtls_base_p99)
  let latency_v2_stats_nullvm_both_p99_trending = constructDataSeries(latency_v2_stats_nullvm_both_p99, latency_none_mtls_base_p99)
  let latency_v2_stats_wasm_both_p99_trending = constructDataSeries(latency_v2_stats_wasm_both_p99, latency_none_mtls_base_p99)
  let latency_v2_sd_nologging_nullvm_both_p99_trending = constructDataSeries(latency_v2_sd_nologging_nullvm_both_p99, latency_none_mtls_base_p99)
  let latency_v2_sd_full_nullvm_both_p99_trending = constructDataSeries(latency_v2_sd_full_nullvm_both_p99, latency_none_mtls_base_p99)

  p99ModesData = [];
  p99ModesData.push(latency_none_mtls_both_p99_trending);
  p99ModesData.push(latency_none_plaintext_both_p99_trending);
  p99ModesData.push(latency_v2_stats_nullvm_both_p99_trending);
  p99ModesData.push(latency_v2_stats_wasm_both_p99_trending);
  p99ModesData.push(latency_v2_sd_nologging_nullvm_both_p99_trending);
  p99ModesData.push(latency_v2_sd_full_nullvm_both_p99_trending);

  generateChart("trending_p99", p99ModesData);
};


