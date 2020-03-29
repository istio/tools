window.onload = function () {
  // p90
  let chart_mixer_both_p90_pattern = constructDataSeries(mixer_both_p90_pattern, none_mtls_base_p90_pattern);
  let chart_none_mtls_both_p90_pattern = constructDataSeries(none_mtls_both_p90_pattern, none_mtls_base_p90_pattern);
  let chart_v2_sd_full_nullvm_both_p90_pattern = constructDataSeries(v2_sd_full_nullvm_both_p90_pattern, none_mtls_base_p90_pattern);

  p90ModesData = [];
  p90ModesData.push(none_mtls_base_p90_pattern);
  p90ModesData.push(chart_mixer_both_p90_pattern);
  p90ModesData.push(chart_none_mtls_both_p90_pattern);
  p90ModesData.push(chart_v2_sd_full_nullvm_both_p90_pattern);

  generateChart("chart_p90", p90ModesData);

  // p99
  let chart_mixer_both_p99_pattern = constructDataSeries(mixer_both_p99_pattern, none_mtls_base_p99_pattern);
  let chart_none_mtls_both_p99_pattern = constructDataSeries(none_mtls_both_p99_pattern, none_mtls_base_p99_pattern);
  let chart_v2_sd_full_nullvm_both_p99_pattern = constructDataSeries(v2_sd_full_nullvm_both_p99_pattern, none_mtls_base_p99_pattern);

  p99ModesData = [];
  p99ModesData.push(none_mtls_base_p99_pattern);
  p99ModesData.push(chart_mixer_both_p99_pattern);
  p99ModesData.push(chart_none_mtls_both_p99_pattern);
  p99ModesData.push(chart_v2_sd_full_nullvm_both_p99_pattern);

  generateChart("chart_p99", p99ModesData);
};


