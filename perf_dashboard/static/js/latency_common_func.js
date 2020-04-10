window.generateLatencyChart = function(connNum, options) {
  // p50-release
  p50ReleaseModesData = [];
  p50ReleaseModesData.push(latency_none_mtls_base_p50);
  p50ReleaseModesData.push(latency_none_mtls_both_p50);
  p50ReleaseModesData.push(latency_none_plaintext_both_p50);
  p50ReleaseModesData.push(latency_v2_stats_nullvm_both_p50);
  p50ReleaseModesData.push(latency_v2_sd_nologging_nullvm_both_p50);
  p50ReleaseModesData.push(latency_v2_sd_full_nullvm_both_p50);

  generateLatencyChartByID("latency-p50-release", connNum, p50ReleaseModesData, options)

  // p90-release
  p90ReleaseModesData = [];
  p90ReleaseModesData.push(latency_none_mtls_base_p90);
  p90ReleaseModesData.push(latency_none_mtls_both_p90);
  p90ReleaseModesData.push(latency_none_plaintext_both_p90);
  p90ReleaseModesData.push(latency_v2_stats_nullvm_both_p90);
  p90ReleaseModesData.push(latency_v2_sd_nologging_nullvm_both_p90);
  p90ReleaseModesData.push(latency_v2_sd_full_nullvm_both_p90);

  generateLatencyChartByID("latency-p90-release", connNum, p90ReleaseModesData, options)


  // p99-release
  p99ReleaseModesData = [];
  p99ReleaseModesData.push(latency_none_mtls_base_p99);
  p99ReleaseModesData.push(latency_none_mtls_both_p99);
  p99ReleaseModesData.push(latency_none_plaintext_both_p99);
  p99ReleaseModesData.push(latency_v2_stats_nullvm_both_p99);
  p99ReleaseModesData.push(latency_v2_sd_nologging_nullvm_both_p99);
  p99ReleaseModesData.push(latency_v2_sd_full_nullvm_both_p99);

  generateLatencyChartByID("latency-p99-release", connNum, p99ReleaseModesData, options)

  // p50-master
  p50ModesData = [];
  p50ModesData.push(latency_none_mtls_base_p50_master);
  p50ModesData.push(latency_none_mtls_both_p50_master);
  p50ModesData.push(latency_none_plaintext_both_p50_master);
  p50ModesData.push(latency_v2_stats_nullvm_both_p50_master);
  p50ModesData.push(latency_v2_sd_nologging_nullvm_both_p50_master);
  p50ModesData.push(latency_v2_sd_full_nullvm_both_p50_master);

  generateLatencyChartByID("latency-p50-master", connNum, p50ModesData, options)

  // p90-master
  p90ModesData = [];
  p90ModesData.push(latency_none_mtls_base_p90_master);
  p90ModesData.push(latency_none_mtls_both_p90_master);
  p90ModesData.push(latency_none_plaintext_both_p90_master);
  p90ModesData.push(latency_v2_stats_nullvm_both_p90_master);
  p90ModesData.push(latency_v2_sd_nologging_nullvm_both_p90_master);
  p90ModesData.push(latency_v2_sd_full_nullvm_both_p90_master);

  generateLatencyChartByID("latency-p90-master", connNum, p90ModesData, options)


  // p99-master
  p99ModesData = [];
  p99ModesData.push(latency_none_mtls_base_p99_master);
  p99ModesData.push(latency_none_mtls_both_p99_master);
  p99ModesData.push(latency_none_plaintext_both_p99_master);
  p99ModesData.push(latency_v2_stats_nullvm_both_p99_master);
  p99ModesData.push(latency_v2_sd_nologging_nullvm_both_p99_master);
  p99ModesData.push(latency_v2_sd_full_nullvm_both_p99_master);

  generateLatencyChartByID("latency-p99-master", connNum, p99ModesData, options)
};

window.generateLatencyChartByID = function(chartID, connNum, modesData, options) {
    new Chart(document.getElementById(chartID), {
    type: 'line',
    data: {
        labels: connNum,
        datasets: [
            {
                label: "baseline",
                backgroundColor: "rgba(236, 66, 53, 0.2)",
                borderColor: "rgba(236, 66, 53, 1)",
                data: modesData[0],
                fill: false
            }, {
                label: "none-mtls_both",
                backgroundColor: "rgba(0, 0, 0, 0.2)",
                borderColor: "rgba(0, 0, 0, 1)",
                data: modesData[1],
                fill: false
            }, {
                label: "none-plaintext_both",
                backgroundColor: "rgba(52, 235, 219, 0.2)",
                borderColor: "rgba(52, 235, 219, 1)",
                data: modesData[2],
                fill: false
            }, {
                label: "v2-stats-nullvm_both",
                backgroundColor: "rgba(252, 123, 3, 0.2)",
                borderColor: "rgba(252, 123, 3, 1)",
                data: modesData[3],
                fill: false
            }, {
                label: "v2-sd-nologging-nullvm_both",
                backgroundColor: "rgba(52, 168, 85, 0.2)",
                borderColor: "rgba(52, 168, 85, 1)",
                data: modesData[4],
                hidden: true,
                fill: false
            }, {
                label: "v2-sd-full-nullvm_both",
                backgroundColor: "rgba(168, 50, 168, 0.2)",
                borderColor: "rgba(168, 50, 168, 1)",
                data: modesData[5],
                hidden: true,
                fill: false
            }
        ]
    },
    options: options
 });
};

