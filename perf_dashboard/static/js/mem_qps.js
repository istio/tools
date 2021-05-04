// labels along the x-axis
var qpsNum = ["10", "100", "200", "400", "800", "1000"];
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
                labelString: "QPS"
            },
        }]
    }
};

new Chart(document.getElementById("mem-client-qps-release"), {
    type: 'line',
    data: convertData({
        labels: qpsNum,
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
                hidden: true,
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
            }, {
                label: "none-security_authz_ip_both",
                backgroundColor: "rgba(50,219,199, 0.2)",
                borderColor: "rgb(50,219,199)",
                data: mem_client_none_security_authz_ip_both,
                hidden: false,
                fill: false
            }, {
                label: "none-security_authz_path_both",
                backgroundColor: "rgba(8,171,195, 0.2)",
                borderColor: "rgb(8,171,195)",
                data: mem_client_none_security_authz_path_both,
                hidden: false,
                fill: false
            }, {
                label: "none-security_authz_jwt_both",
                backgroundColor: "rgba(5,140,212, 0.2)",
                borderColor: "rgb(5,140,212)",
                data: mem_client_none_security_authz_jwt_both,
                hidden: false,
                fill: false
            }, {
                label: "none-security_peer_authn_both",
                backgroundColor: "rgba(6,91,184, 0.2)",
                borderColor: "rgb(6,91,184)",
                data: mem_client_none_security_peer_authn_both,
                hidden: false,
                fill: false
            }
        ]
    }),
    options: memOptions
});

new Chart(document.getElementById("mem-client-qps-master"), {
    type: 'line',
    data: convertData({
        labels: qpsNum,
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
                hidden: true,
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
            }, {
                label: "none-security_authz_ip_both",
                backgroundColor: "rgba(50,219,199, 0.2)",
                borderColor: "rgb(50,219,199)",
                data: mem_client_none_security_authz_ip_both_master,
                hidden: false,
                fill: false
            }, {
                label: "none-security_authz_path_both",
                backgroundColor: "rgba(8,171,195, 0.2)",
                borderColor: "rgb(8,171,195)",
                data: mem_client_none_security_authz_path_both_master,
                hidden: false,
                fill: false
            }, {
                label: "none-security_authz_jwt_both",
                backgroundColor: "rgba(5,140,212, 0.2)",
                borderColor: "rgb(5,140,212)",
                data: mem_client_none_security_authz_jwt_both_master,
                hidden: false,
                fill: false
            }, {
                label: "none-security_peer_authn_both",
                backgroundColor: "rgba(6,91,184, 0.2)",
                borderColor: "rgb(6,91,184)",
                data: mem_client_none_security_peer_authn_both_master,
                hidden: false,
                fill: false
            }
        ]
    }),
    options: memOptions
});

new Chart(document.getElementById("mem-server-qps-release"), {
    type: 'line',
    data: convertData({
        labels: qpsNum,
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
                hidden: true,
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
            }, {
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
            }, {
                label: "none-security_authz_ip_both",
                backgroundColor: "rgba(50,219,199, 0.2)",
                borderColor: "rgb(50,219,199)",
                data: mem_server_none_security_authz_ip_both,
                hidden: false,
                fill: false
            }, {
                label: "none-security_authz_path_both",
                backgroundColor: "rgba(8,171,195, 0.2)",
                borderColor: "rgb(8,171,195)",
                data: mem_server_none_security_authz_path_both,
                hidden: false,
                fill: false
            }, {
                label: "none-security_authz_jwt_both",
                backgroundColor: "rgba(5,140,212, 0.2)",
                borderColor: "rgb(5,140,212)",
                data: mem_server_none_security_authz_jwt_both,
                hidden: false,
                fill: false
            }, {
                label: "none-security_peer_authn_both",
                backgroundColor: "rgba(6,91,184, 0.2)",
                borderColor: "rgb(6,91,184)",
                data: mem_server_none_security_peer_authn_both,
                hidden: false,
                fill: false
            }
        ]
    }),
    options: memOptions
});

new Chart(document.getElementById("mem-server-qps-master"), {
    type: 'line',
    data: convertData({
        labels: qpsNum,
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
                hidden: true,
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
            }, {
                label: "none-security_authz_ip_both",
                backgroundColor: "rgba(50,219,199, 0.2)",
                borderColor: "rgb(50,219,199)",
                data: mem_server_none_security_authz_ip_both_master,
                hidden: false,
                fill: false
            }, {
                label: "none-security_authz_path_both",
                backgroundColor: "rgba(8,171,195, 0.2)",
                borderColor: "rgb(8,171,195)",
                data: mem_server_none_security_authz_path_both_master,
                hidden: false,
                fill: false
            }, {
                label: "none-security_authz_jwt_both",
                backgroundColor: "rgba(5,140,212, 0.2)",
                borderColor: "rgb(5,140,212)",
                data: mem_server_none_security_authz_jwt_both_master,
                hidden: false,
                fill: false
            }, {
                label: "none-security_peer_authn_both",
                backgroundColor: "rgba(6,91,184, 0.2)",
                borderColor: "rgb(6,91,184)",
                data: mem_server_none_security_peer_authn_both_master,
                hidden: false,
                fill: false
            }
        ]
    }),
    options: memOptions
});

new Chart(document.getElementById("mem-ingressgw-qps-release"), {
    type: 'line',
    data: convertData({
        labels: qpsNum,
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
                hidden: true,
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
            }, {
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

new Chart(document.getElementById("mem-ingressgw-qps-master"), {
    type: 'line',
    data: convertData({
        labels: qpsNum,
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
                hidden: true,
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
