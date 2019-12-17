window.onload = function () {
  // p90
  let chart_master_pattern_mixer_serveronly_p90 = constructDataSeries(master_pattern_mixer_serveronly_p90, master_pattern_mixer_base_p90);
  let chart_master_pattern_mixer_both_p90 = constructDataSeries(master_pattern_mixer_both_p90, master_pattern_mixer_base_p90);
  let chart_master_pattern_nomixer_serveronly_p90 = constructDataSeries(master_pattern_nomixer_serveronly_p90, master_pattern_mixer_base_p90);
  let chart_master_pattern_nomixer_both_p90 = constructDataSeries(master_pattern_nomixer_both_p90, master_pattern_mixer_base_p90);
  let chart_master_pattern_v2_serveronly_p90 = constructDataSeries(master_pattern_v2_serveronly_p90, master_pattern_mixer_base_p90);
  let chart_master_pattern_v2_both_p90 = constructDataSeries(master_pattern_v2_both_p90, master_pattern_mixer_base_p90);

  p90ModesData = [];
  p90ModesData.push(chart_master_pattern_mixer_serveronly_p90);
  p90ModesData.push(chart_master_pattern_mixer_both_p90);
  p90ModesData.push(chart_master_pattern_nomixer_serveronly_p90);
  p90ModesData.push(chart_master_pattern_nomixer_both_p90);
  p90ModesData.push(chart_master_pattern_v2_serveronly_p90);
  p90ModesData.push(chart_master_pattern_v2_both_p90);

  generateChart("chart_p90_master", p90ModesData);

  // p99
  let chart_master_pattern_mixer_serveronly_p99 = constructDataSeries(master_pattern_mixer_serveronly_p99, master_pattern_mixer_base_p99);
  let chart_master_pattern_mixer_both_p99 = constructDataSeries(master_pattern_mixer_both_p99, master_pattern_mixer_base_p99);
  let chart_master_pattern_nomixer_serveronly_p99 = constructDataSeries(master_pattern_nomixer_serveronly_p99, master_pattern_mixer_base_p99);
  let chart_master_pattern_nomixer_both_p99 = constructDataSeries(master_pattern_nomixer_both_p99, master_pattern_mixer_base_p99);
  let chart_master_pattern_v2_serveronly_p99 = constructDataSeries(master_pattern_v2_serveronly_p99, master_pattern_mixer_base_p99);
  let chart_master_pattern_v2_both_p99 = constructDataSeries(master_pattern_v2_both_p99, master_pattern_mixer_base_p99);

  p99ModesData = [];
  p99ModesData.push(chart_master_pattern_mixer_serveronly_p99);
  p99ModesData.push(chart_master_pattern_mixer_both_p99);
  p99ModesData.push(chart_master_pattern_nomixer_serveronly_p99);
  p99ModesData.push(chart_master_pattern_nomixer_both_p99);
  p99ModesData.push(chart_master_pattern_v2_serveronly_p99);
  p99ModesData.push(chart_master_pattern_v2_both_p99);

  generateChart("chart_p99_master", p99ModesData);
};


