window.onload = function () {
  // p90
  let chart_none_mtls_both_p90_pattern_master = constructDataSeries(none_mtls_both_p90_pattern_master, none_mtls_base_p90_pattern_master);
  let chart_v2_sd_full_nullvm_both_p90_pattern_master = constructDataSeries(v2_sd_full_nullvm_both_p90_pattern_master, none_mtls_base_p90_pattern_master);

  p90ModesData = [];
  p90ModesData.push(chart_none_mtls_both_p90_pattern_master);
  p90ModesData.push(chart_v2_sd_full_nullvm_both_p90_pattern_master);

  generateChart("chart_p90_master", p90ModesData);

  // p99
  let chart_none_mtls_both_p99_pattern_master = constructDataSeries(none_mtls_both_p99_pattern_master, none_mtls_base_p99_pattern_master);
  let chart_v2_sd_full_nullvm_both_p99_pattern_master = constructDataSeries(v2_sd_full_nullvm_both_p99_pattern_master, none_mtls_base_p99_pattern_master);

  p99ModesData = [];
  p99ModesData.push(chart_none_mtls_both_p99_pattern_master);
  p99ModesData.push(chart_v2_sd_full_nullvm_both_p99_pattern_master);

  generateChart("chart_p99_master", p99ModesData);
};


