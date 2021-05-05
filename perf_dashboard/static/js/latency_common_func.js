window.generateLatencyChart = function(xNum, options) {
  // p50-release
  p50ReleaseModesData = [];
  p50ReleaseModesData.push(latency_none_mtls_base_p50);
  p50ReleaseModesData.push(latency_none_mtls_both_p50);
  p50ReleaseModesData.push(latency_none_plaintext_both_p50);
  p50ReleaseModesData.push(latency_v2_stats_nullvm_both_p50);
  p50ReleaseModesData.push(latency_v2_stats_wasm_both_p50);
  p50ReleaseModesData.push(latency_v2_sd_nologging_nullvm_both_p50);
  p50ReleaseModesData.push(latency_v2_sd_full_nullvm_both_p50);
  p50ReleaseModesData.push(latency_none_security_authz_ip_both_p50);
  p50ReleaseModesData.push(latency_none_security_authz_path_both_p50);
  p50ReleaseModesData.push(latency_none_security_authz_jwt_both_p50);
  p50ReleaseModesData.push(latency_none_security_peer_authn_both_p50);

  generateLatencyChartByID("latency-p50-release", xNum, p50ReleaseModesData, options)

  // p90-release
  p90ReleaseModesData = [];
  p90ReleaseModesData.push(latency_none_mtls_base_p90);
  p90ReleaseModesData.push(latency_none_mtls_both_p90);
  p90ReleaseModesData.push(latency_none_plaintext_both_p90);
  p90ReleaseModesData.push(latency_v2_stats_nullvm_both_p90);
  p90ReleaseModesData.push(latency_v2_stats_wasm_both_p90);
  p90ReleaseModesData.push(latency_v2_sd_nologging_nullvm_both_p90);
  p90ReleaseModesData.push(latency_v2_sd_full_nullvm_both_p90);
  p90ReleaseModesData.push(latency_none_security_authz_ip_both_p90);
  p90ReleaseModesData.push(latency_none_security_authz_path_both_p90);
  p90ReleaseModesData.push(latency_none_security_authz_jwt_both_p90);
  p90ReleaseModesData.push(latency_none_security_peer_authn_both_p90);

  generateLatencyChartByID("latency-p90-release", xNum, p90ReleaseModesData, options)

  // p99-release
  p99ReleaseModesData = [];
  p99ReleaseModesData.push(latency_none_mtls_base_p99);
  p99ReleaseModesData.push(latency_none_mtls_both_p99);
  p99ReleaseModesData.push(latency_none_plaintext_both_p99);
  p99ReleaseModesData.push(latency_v2_stats_nullvm_both_p99);
  p99ReleaseModesData.push(latency_v2_stats_wasm_both_p99);
  p99ReleaseModesData.push(latency_v2_sd_nologging_nullvm_both_p99);
  p99ReleaseModesData.push(latency_v2_sd_full_nullvm_both_p99);
  p99ReleaseModesData.push(latency_none_security_authz_ip_both_p99);
  p99ReleaseModesData.push(latency_none_security_authz_path_both_p99);
  p99ReleaseModesData.push(latency_none_security_authz_jwt_both_p99);
  p99ReleaseModesData.push(latency_none_security_peer_authn_both_p99);

  generateLatencyChartByID("latency-p99-release", xNum, p99ReleaseModesData, options)

  // p99.9-release
  p999ReleaseModesData = [];
  p999ReleaseModesData.push(latency_none_mtls_base_p999);
  p999ReleaseModesData.push(latency_none_mtls_both_p999);
  p999ReleaseModesData.push(latency_none_plaintext_both_p999);
  p999ReleaseModesData.push(latency_v2_stats_nullvm_both_p999);
  p999ReleaseModesData.push(latency_v2_stats_wasm_both_p999);
  p999ReleaseModesData.push(latency_v2_sd_nologging_nullvm_both_p999);
  p999ReleaseModesData.push(latency_v2_sd_full_nullvm_both_p999);
  p999ReleaseModesData.push(latency_none_security_authz_ip_both_p999);
  p999ReleaseModesData.push(latency_none_security_authz_path_both_p999);
  p999ReleaseModesData.push(latency_none_security_authz_jwt_both_p999);
  p999ReleaseModesData.push(latency_none_security_peer_authn_both_p999);

  generateLatencyChartByID("latency-p999-release", xNum, p999ReleaseModesData, options)

  // p50-master
  p50ModesData = [];
  p50ModesData.push(latency_none_mtls_base_p50_master);
  p50ModesData.push(latency_none_mtls_both_p50_master);
  p50ModesData.push(latency_none_plaintext_both_p50_master);
  p50ModesData.push(latency_v2_stats_nullvm_both_p50_master);
  p50ModesData.push(latency_v2_stats_wasm_both_p50_master);
  p50ModesData.push(latency_v2_sd_nologging_nullvm_both_p50_master);
  p50ModesData.push(latency_v2_sd_full_nullvm_both_p50_master);
  p50ModesData.push(latency_none_security_authz_ip_both_p50_master);
  p50ModesData.push(latency_none_security_authz_path_both_p50_master);
  p50ModesData.push(latency_none_security_authz_jwt_both_p50_master);
  p50ModesData.push(latency_none_security_peer_authn_both_p50_master);

  generateLatencyChartByID("latency-p50-master", xNum, p50ModesData, options)

  // p90-master
  p90ModesData = [];
  p90ModesData.push(latency_none_mtls_base_p90_master);
  p90ModesData.push(latency_none_mtls_both_p90_master);
  p90ModesData.push(latency_none_plaintext_both_p90_master);
  p90ModesData.push(latency_v2_stats_nullvm_both_p90_master);
  p90ModesData.push(latency_v2_stats_wasm_both_p90_master);
  p90ModesData.push(latency_v2_sd_nologging_nullvm_both_p90_master);
  p90ModesData.push(latency_v2_sd_full_nullvm_both_p90_master);
  p90ModesData.push(latency_none_security_authz_ip_both_p90_master);
  p90ModesData.push(latency_none_security_authz_path_both_p90_master);
  p90ModesData.push(latency_none_security_authz_jwt_both_p90_master);
  p90ModesData.push(latency_none_security_peer_authn_both_p90_master);

  generateLatencyChartByID("latency-p90-master", xNum, p90ModesData, options)

  // p99-master
  p99ModesData = [];
  p99ModesData.push(latency_none_mtls_base_p99_master);
  p99ModesData.push(latency_none_mtls_both_p99_master);
  p99ModesData.push(latency_none_plaintext_both_p99_master);
  p99ModesData.push(latency_v2_stats_nullvm_both_p99_master);
  p99ModesData.push(latency_v2_stats_wasm_both_p99_master);
  p99ModesData.push(latency_v2_sd_nologging_nullvm_both_p99_master);
  p99ModesData.push(latency_v2_sd_full_nullvm_both_p99_master);
  p99ModesData.push(latency_none_security_authz_ip_both_p99_master);
  p99ModesData.push(latency_none_security_authz_path_both_p99_master);
  p99ModesData.push(latency_none_security_authz_jwt_both_p99_master);
  p99ModesData.push(latency_none_security_peer_authn_both_p99_master);

  generateLatencyChartByID("latency-p99-master", xNum, p99ModesData, options)

  // p99.9-master
  p999ModesData = [];
  p999ModesData.push(latency_none_mtls_base_p999_master);
  p999ModesData.push(latency_none_mtls_both_p999_master);
  p999ModesData.push(latency_none_plaintext_both_p999_master);
  p999ModesData.push(latency_v2_stats_nullvm_both_p999_master);
  p999ModesData.push(latency_v2_stats_wasm_both_p999_master);
  p999ModesData.push(latency_v2_sd_nologging_nullvm_both_p999_master);
  p999ModesData.push(latency_v2_sd_full_nullvm_both_p999_master);
  p999ModesData.push(latency_none_security_authz_ip_both_p999_master);
  p999ModesData.push(latency_none_security_authz_path_both_p999_master);
  p999ModesData.push(latency_none_security_authz_jwt_both_p999_master);
  p999ModesData.push(latency_none_security_peer_authn_both_p999_master);

  generateLatencyChartByID("latency-p999-master", xNum, p999ModesData, options)
};

window.generateLatencyChartByID = function(chartID, xNum, modesData, options) {
    new Chart(document.getElementById(chartID), {
    type: 'line',
    data: convertData({
        labels: xNum,
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
                hidden: true,
                fill: false
            }, {
                label: "v2-stats-nullvm_both",
                backgroundColor: "rgba(252, 123, 3, 0.2)",
                borderColor: "rgba(252, 123, 3, 1)",
                data: modesData[3],
                fill: false
            }, {
                label: "v2-stats-wasm_both",
                backgroundColor: "rgba(242, 245, 66, 0.2)",
                borderColor: "rgba(242, 245, 66, 1)",
                data: modesData[4],
                fill: false
            }, {
                label: "v2-sd-nologging-nullvm_both",
                backgroundColor: "rgba(52, 168, 85, 0.2)",
                borderColor: "rgba(52, 168, 85, 1)",
                data: modesData[5],
                hidden: true,
                fill: false
            }, {
                label: "v2-sd-full-nullvm_both",
                backgroundColor: "rgba(168, 50, 168, 0.2)",
                borderColor: "rgba(168, 50, 168, 1)",
                data: modesData[6],
                hidden: true,
                fill: false
            }, {
                label: "none-security_authz_ip_both",
                backgroundColor: "rgba(50,219,199, 0.2)",
                borderColor: "rgb(50,219,199)",
                data: modesData[7],
                hidden: false,
                fill: false
            }, {
                label: "none-security_authz_path_both",
                backgroundColor: "rgba(8,171,195, 0.2)",
                borderColor: "rgb(8,171,195)",
                data: modesData[8],
                hidden: false,
                fill: false
            }, {
                label: "none-security_authz_jwt_both",
                backgroundColor: "rgba(5,140,212, 0.2)",
                borderColor: "rgb(5,140,212)",
                data: modesData[9],
                hidden: false,
                fill: false
            }, {
                label: "none-security_peer_authn_both",
                backgroundColor: "rgba(6,91,184, 0.2)",
                borderColor: "rgb(6,91,184)",
                data: modesData[10],
                hidden: false,
                fill: false
            }
        ]
    }),
    options: options
 });
};

