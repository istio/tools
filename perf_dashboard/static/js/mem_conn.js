// labels along the x-axis
var connNum = ["2", "4", "8", "16", "32", "64"];
// x-axis and y-axis labels
var memOptions = {
    scales: {
        yAxes: [{
            scaleLabel: {
                display: true,
                labelString: "istio-proxy average Memory (Mi)"
            },
        }],
        xAxes: [{
            type:"linear",
            scaleLabel: {
                display: true,
                labelString: "Connections"
            },
        }]
    }
};

new Chart(document.getElementById("mem-client-conn-release"), {
    type: 'line',
    data: convertData({
        labels: connNum,
        datasets: [
            {
                label: "baseline",
                backgroundColor: "rgba(236, 66, 53, 0.2)",
                borderColor: "rgba(236, 66, 53, 1)",
                data: mem_client_none_mtls_base,
                fill: false
            }, {
                label: "none-mtls_both",
                backgroundColor: "rgba(0, 0, 0, 0.2)",
                borderColor: "rgba(0, 0, 0, 1)",
                data: mem_client_none_mtls_both,
                fill: false
            },{
                label: "none-plaintext_both",
                backgroundColor: "rgba(52, 235, 219, 0.2)",
                borderColor: "rgba(52, 235, 219, 1)",
                data: mem_client_none_plaintext_both,
                fill: false
            }, {
                label: "v2-stats-nullvm_both",
                backgroundColor: "rgba(252, 123, 3, 0.2)",
                borderColor: "rgba(252, 123, 3, 1)",
                data: mem_client_v2_stats_nullvm_both,
                fill: false
            }, {
                label: "v2-stats-wasm_both",
                backgroundColor: "rgba(242, 245, 66, 0.2)",
                borderColor: "rgba(242, 245, 66, 1)",
                data: mem_client_v2_stats_wasm_both,
                fill: false
            }, {
                label: "v2-sd-nologging-nullvm_both",
                backgroundColor: "rgba(52, 168, 85, 0.2)",
                borderColor: "rgba(52, 168, 85, 1)",
                data: mem_client_v2_sd_nologging_nullvm_both,
                hidden: true,
                fill: false
            }, {
                label: "v2-sd-full-nullvm_both",
                backgroundColor: "rgba(168, 50, 168, 0.2)",
                borderColor: "rgba(168, 50, 168, 1)",
                data: mem_client_v2_sd_full_nullvm_both,
                hidden: true,
                fill: false
            }
        ]
    }),
    options: memOptions
});

new Chart(document.getElementById("mem-client-conn-master"), {
    type: 'line',
    data: convertData({
        labels: connNum,
        datasets: [
            {
                label: "baseline",
                backgroundColor: "rgba(236, 66, 53, 0.2)",
                borderColor: "rgba(236, 66, 53, 1)",
                data: mem_client_none_mtls_base_master,
                fill: false
            }, {
                label: "none-mtls_both",
                backgroundColor: "rgba(0, 0, 0, 0.2)",
                borderColor: "rgba(0, 0, 0, 1)",
                data: mem_client_none_mtls_both_master,
                fill: false
            },{
                label: "none-plaintext_both",
                backgroundColor: "rgba(52, 235, 219, 0.2)",
                borderColor: "rgba(52, 235, 219, 1)",
                data: mem_client_none_plaintext_both_master,
                fill: false
            }, {
                label: "v2-stats-nullvm_both",
                backgroundColor: "rgba(252, 123, 3, 0.2)",
                borderColor: "rgba(252, 123, 3, 1)",
                data: mem_client_v2_stats_nullvm_both_master,
                fill: false
            }, {
                label: "v2-stats-wasm_both",
                backgroundColor: "rgba(242, 245, 66, 0.2)",
                borderColor: "rgba(242, 245, 66, 1)",
                data: mem_client_v2_stats_wasm_both_master,
                fill: false
            }, {
                label: "v2-sd-nologging-nullvm_both",
                backgroundColor: "rgba(52, 168, 85, 0.2)",
                borderColor: "rgba(52, 168, 85, 1)",
                data: mem_client_v2_sd_nologging_nullvm_both_master,
                hidden: true,
                fill: false
            }, {
                label: "v2-sd-full-nullvm_both",
                backgroundColor: "rgba(168, 50, 168, 0.2)",
                borderColor: "rgba(168, 50, 168, 1)",
                data: mem_client_v2_sd_full_nullvm_both_master,
                hidden: true,
                fill: false
            }
        ]
    }),
    options: memOptions
});

new Chart(document.getElementById("mem-server-conn-release"), {
    type: 'line',
    data: convertData({
        labels: connNum,
        datasets: [
            {
                label: "baseline",
                backgroundColor: "rgba(236, 66, 53, 0.2)",
                borderColor: "rgba(236, 66, 53, 1)",
                data: mem_server_none_mtls_base,
                fill: false
            }, {
                label: "none-mtls_both",
                backgroundColor: "rgba(0, 0, 0, 0.2)",
                borderColor: "rgba(0, 0, 0, 1)",
                data: mem_server_none_mtls_both,
                fill: false
            }, {
                label: "none-plaintext_both",
                backgroundColor: "rgba(52, 235, 219, 0.2)",
                borderColor: "rgba(52, 235, 219, 1)",
                data: mem_server_none_plaintext_both,
                fill: false
            }, {
                label: "v2-stats-nullvm_both",
                backgroundColor: "rgba(252, 123, 3, 0.2)",
                borderColor: "rgba(252, 123, 3, 1)",
                data: mem_server_v2_stats_nullvm_both,
                fill: false
            }, {
                label: "v2-stats-wasm_both",
                backgroundColor: "rgba(242, 245, 66, 0.2)",
                borderColor: "rgba(242, 245, 66, 1)",
                data: mem_server_v2_stats_wasm_both,
                fill: false
            },{
                label: "v2-sd-nologging-nullvm_both",
                backgroundColor: "rgba(52, 168, 85, 0.2)",
                borderColor: "rgba(52, 168, 85, 1)",
                data: mem_server_v2_sd_nologging_nullvm_both,
                hidden: true,
                fill: false
            }, {
                label: "v2-sd-full-nullvm_both",
                backgroundColor: "rgba(168, 50, 168, 0.2)",
                borderColor: "rgba(168, 50, 168, 1)",
                data: mem_server_v2_sd_full_nullvm_both,
                hidden: true,
                fill: false
            }
        ]
    }),
    options: memOptions
});

new Chart(document.getElementById("mem-server-conn-master"), {
    type: 'line',
    data: convertData({
        labels: connNum,
        datasets: [
            {
                label: "baseline",
                backgroundColor: "rgba(236, 66, 53, 0.2)",
                borderColor: "rgba(236, 66, 53, 1)",
                data: mem_server_none_mtls_base_master,
                fill: false
            }, {
                label: "none-mtls_both",
                backgroundColor: "rgba(0, 0, 0, 0.2)",
                borderColor: "rgba(0, 0, 0, 1)",
                data: mem_server_none_mtls_both_master,
                fill: false
            }, {
                label: "none-plaintext_both",
                backgroundColor: "rgba(52, 235, 219, 0.2)",
                borderColor: "rgba(52, 235, 219, 1)",
                data: mem_server_none_plaintext_both_master,
                fill: false
            }, {
                label: "v2-stats-nullvm_both",
                backgroundColor: "rgba(252, 123, 3, 0.2)",
                borderColor: "rgba(252, 123, 3, 1)",
                data: mem_server_v2_stats_nullvm_both_master,
                fill: false
            }, {
                label: "v2-stats-wasm_both",
                backgroundColor: "rgba(242, 245, 66, 0.2)",
                borderColor: "rgba(242, 245, 66, 1)",
                data: mem_server_v2_stats_wasm_both_master,
                fill: false
            }, {
                label: "v2-sd-nologging-nullvm_both",
                backgroundColor: "rgba(52, 168, 85, 0.2)",
                borderColor: "rgba(52, 168, 85, 1)",
                data: mem_server_v2_sd_nologging_nullvm_both_master,
                hidden: true,
                fill: false
            }, {
                label: "v2-sd-full-nullvm_both",
                backgroundColor: "rgba(168, 50, 168, 0.2)",
                borderColor: "rgba(168, 50, 168, 1)",
                data: mem_server_v2_sd_full_nullvm_both_master,
                hidden: true,
                fill: false
            }
        ]
    }),
    options: memOptions
});

new Chart(document.getElementById("mem-ingressgw-conn-release"), {
    type: 'line',
    data: convertData({
        labels: connNum,
        datasets: [
            {
                label: "baseline",
                backgroundColor: "rgba(236, 66, 53, 0.2)",
                borderColor: "rgba(236, 66, 53, 1)",
                data: mem_ingressgw_none_mtls_base,
                fill: false
            }, {
                label: "none-mtls_both",
                backgroundColor: "rgba(0, 0, 0, 0.2)",
                borderColor: "rgba(0, 0, 0, 1)",
                data: mem_ingressgw_none_mtls_both,
                fill: false
            }, {
                label: "none-plaintext_both",
                backgroundColor: "rgba(52, 235, 219, 0.2)",
                borderColor: "rgba(52, 235, 219, 1)",
                data: mem_ingressgw_none_plaintext_both,
                fill: false
            }, {
                label: "v2-stats-nullvm_both",
                backgroundColor: "rgba(252, 123, 3, 0.2)",
                borderColor: "rgba(252, 123, 3, 1)",
                data: mem_ingressgw_v2_stats_nullvm_both,
                fill: false
            }, {
                label: "v2-stats-wasm_both",
                backgroundColor: "rgba(242, 245, 66, 0.2)",
                borderColor: "rgba(242, 245, 66, 1)",
                data: mem_ingressgw_v2_stats_wasm_both,
                fill: false
            },{
                label: "v2-sd-nologging-nullvm_both",
                backgroundColor: "rgba(52, 168, 85, 0.2)",
                borderColor: "rgba(52, 168, 85, 1)",
                data: mem_ingressgw_v2_sd_nologging_nullvm_both,
                hidden: true,
                fill: false
            }, {
                label: "v2-sd-full-nullvm_both",
                backgroundColor: "rgba(168, 50, 168, 0.2)",
                borderColor: "rgba(168, 50, 168, 1)",
                data: mem_ingressgw_v2_sd_full_nullvm_both,
                hidden: true,
                fill: false
            }
        ]
    }),
    options: memOptions
});

new Chart(document.getElementById("mem-ingressgw-conn-master"), {
    type: 'line',
    data: convertData({
        labels: connNum,
        datasets: [
            {
                label: "baseline",
                backgroundColor: "rgba(236, 66, 53, 0.2)",
                borderColor: "rgba(236, 66, 53, 1)",
                data: mem_ingressgw_none_mtls_base_master,
                fill: false
            }, {
                label: "none-mtls_both",
                backgroundColor: "rgba(0, 0, 0, 0.2)",
                borderColor: "rgba(0, 0, 0, 1)",
                data: mem_ingressgw_none_mtls_both_master,
                fill: false
            }, {
                label: "none-plaintext_both",
                backgroundColor: "rgba(52, 235, 219, 0.2)",
                borderColor: "rgba(52, 235, 219, 1)",
                data: mem_ingressgw_none_plaintext_both_master,
                fill: false
            }, {
                label: "v2-stats-nullvm_both",
                backgroundColor: "rgba(252, 123, 3, 0.2)",
                borderColor: "rgba(252, 123, 3, 1)",
                data: mem_ingressgw_v2_stats_nullvm_both_master,
                fill: false
            }, {
                label: "v2-stats-wasm_both",
                backgroundColor: "rgba(242, 245, 66, 0.2)",
                borderColor: "rgba(242, 245, 66, 1)",
                data: mem_ingressgw_v2_stats_wasm_both_master,
                fill: false
            }, {
                label: "v2-sd-nologging-nullvm_both",
                backgroundColor: "rgba(52, 168, 85, 0.2)",
                borderColor: "rgba(52, 168, 85, 1)",
                data: mem_ingressgw_v2_sd_nologging_nullvm_both_master,
                hidden: true,
                fill: false
            }, {
                label: "v2-sd-full-nullvm_both",
                backgroundColor: "rgba(168, 50, 168, 0.2)",
                borderColor: "rgba(168, 50, 168, 1)",
                data: mem_ingressgw_v2_sd_full_nullvm_both_master,
                hidden: true,
                fill: false
            }
        ]
    }),
    options: memOptions
});
