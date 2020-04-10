// labels along the x-axis
var qpsNum = ["10", "100", "500", "1000", "2000", "3000"];
// x-axis and y-axis labels
var cpuOptions = {
    scales: {
        yAxes: [{
            scaleLabel: {
                display: true,
                labelString: "max CPUs, server proxy (millicores)"
            }
        }],
        xAxes: [{
            scaleLabel: {
                display: true,
                labelString: "QPS"
            }
        }]
    }
};

var memOptions = {
    scales: {
        yAxes: [{
            scaleLabel: {
                display: true,
                labelString: "max memory usage, server proxy (MB)"
            }
        }],
        xAxes: [{
            scaleLabel: {
                display: true,
                labelString: "QPS"
            }
        }]
    }
};

new Chart(document.getElementById("cpu-qps-release"), {
    type: 'line',
    data: {
        labels: qpsNum,
        datasets: [
            {
                label: "baseline",
                backgroundColor: "rgba(236, 66, 53, 0.2)",
                borderColor: "rgba(236, 66, 53, 1)",
                data: cpu_none_mtls_base,
                fill: false
            }, {
                label: "none-mtls_both",
                backgroundColor: "rgba(0, 0, 0, 0.2)",
                borderColor: "rgba(0, 0, 0, 1)",
                data: cpu_none_mtls_both,
                fill: false
            },{
                label: "none-plaintext_both",
                backgroundColor: "rgba(52, 235, 219, 0.2)",
                borderColor: "rgba(52, 235, 219, 1)",
                data: cpu_none_plaintext_both,
                fill: false
            }, {
                label: "v2-stats-nullvm_both",
                backgroundColor: "rgba(252, 123, 3, 0.2)",
                borderColor: "rgba(252, 123, 3, 1)",
                data: cpu_v2_stats_nullvm_both,
                fill: false
            }, {
                label: "v2-sd-nologging-nullvm_both",
                backgroundColor: "rgba(52, 168, 85, 0.2)",
                borderColor: "rgba(52, 168, 85, 1)",
                data: cpu_v2_sd_nologging_nullvm_both,
                hidden: true,
                fill: false
            }, {
                label: "v2-sd-full-nullvm_both",
                backgroundColor: "rgba(168, 50, 168, 0.2)",
                borderColor: "rgba(168, 50, 168, 1)",
                data: cpu_v2_sd_full_nullvm_both,
                hidden: true,
                fill: false
            }
        ]
    },
    options: cpuOptions
});

new Chart(document.getElementById("mem-qps-release"), {
    type: 'line',
    data: {
        labels: qpsNum,
        datasets: [
            {
                label: "baseline",
                backgroundColor: "rgba(236, 66, 53, 0.2)",
                borderColor: "rgba(236, 66, 53, 1)",
                data: mem_none_mtls_base,
                fill: false
            }, {
                label: "none-mtls_both",
                backgroundColor: "rgba(0, 0, 0, 0.2)",
                borderColor: "rgba(0, 0, 0, 1)",
                data: mem_none_mtls_both,
                fill: false
            }, {
                label: "none-plaintext_both",
                backgroundColor: "rgba(52, 235, 219, 0.2)",
                borderColor: "rgba(52, 235, 219, 1)",
                data: mem_none_plaintext_both,
                fill: false
            }, {
                label: "v2-stats-nullvm_both",
                backgroundColor: "rgba(252, 123, 3, 0.2)",
                borderColor: "rgba(252, 123, 3, 1)",
                data: mem_v2_stats_nullvm_both,
                fill: false
            }, {
                label: "v2-sd-nologging-nullvm_both",
                backgroundColor: "rgba(52, 168, 85, 0.2)",
                borderColor: "rgba(52, 168, 85, 1)",
                data: mem_v2_sd_nologging_nullvm_both,
                hidden: true,
                fill: false
            }, {
                label: "v2-sd-full-nullvm_both",
                backgroundColor: "rgba(168, 50, 168, 0.2)",
                borderColor: "rgba(168, 50, 168, 1)",
                data: mem_v2_sd_full_nullvm_both,
                hidden: true,
                fill: false
            }
        ]
    },
    options: memOptions
});

new Chart(document.getElementById("cpu-qps-master"), {
    type: 'line',
    data: {
        labels: qpsNum,
        datasets: [
            {
                label: "baseline",
                backgroundColor: "rgba(236, 66, 53, 0.2)",
                borderColor: "rgba(236, 66, 53, 1)",
                data: cpu_none_mtls_base_master,
                fill: false
            }, {
                label: "none-mtls_both",
                backgroundColor: "rgba(0, 0, 0, 0.2)",
                borderColor: "rgba(0, 0, 0, 1)",
                data: cpu_none_mtls_both_master,
                fill: false
            }, {
                label: "none-plaintext_both",
                backgroundColor: "rgba(52, 235, 219, 0.2)",
                borderColor: "rgba(52, 235, 219, 1)",
                data: cpu_none_plaintext_both_master,
                fill: false
            }, {
                label: "v2-stats-nullvm_both",
                backgroundColor: "rgba(252, 123, 3, 0.2)",
                borderColor: "rgba(252, 123, 3, 1)",
                data: cpu_v2_stats_nullvm_both_master,
                fill: false
            }, {
                label: "v2-sd-nologging-nullvm_both",
                backgroundColor: "rgba(52, 168, 85, 0.2)",
                borderColor: "rgba(52, 168, 85, 1)",
                data: cpu_v2_sd_nologging_nullvm_both_master,
                hidden: true,
                fill: false
            }, {
                label: "v2-sd-full-nullvm_both",
                backgroundColor: "rgba(168, 50, 168, 0.2)",
                borderColor: "rgba(168, 50, 168, 1)",
                data: cpu_v2_sd_full_nullvm_both_master,
                hidden: true,
                fill: false
            }
        ]
    },
    options: cpuOptions
});

new Chart(document.getElementById("mem-qps-master"), {
    type: 'line',
    data: {
        labels: qpsNum,
        datasets: [
            {
                label: "baseline",
                backgroundColor: "rgba(236, 66, 53, 0.2)",
                borderColor: "rgba(236, 66, 53, 1)",
                data: mem_none_mtls_base_master,
                fill: false
            }, {
                label: "none-mtls_both",
                backgroundColor: "rgba(0, 0, 0, 0.2)",
                borderColor: "rgba(0, 0, 0, 1)",
                data: mem_none_mtls_both_master,
                fill: false
            }, {
                label: "none-plaintext_both",
                backgroundColor: "rgba(52, 235, 219, 0.2)",
                borderColor: "rgba(52, 235, 219, 1)",
                data: mem_none_plaintext_both_master,
                fill: false
            }, {
                label: "v2-stats-nullvm_both",
                backgroundColor: "rgba(252, 123, 3, 0.2)",
                borderColor: "rgba(252, 123, 3, 1)",
                data: mem_v2_stats_nullvm_both_master,
                fill: false
            }, {
                label: "v2-sd-nologging-nullvm_both",
                backgroundColor: "rgba(52, 168, 85, 0.2)",
                borderColor: "rgba(52, 168, 85, 1)",
                data: mem_v2_sd_nologging_nullvm_both_master,
                hidden: true,
                fill: false
            }, {
                label: "v2-sd-full-nullvm_both",
                backgroundColor: "rgba(168, 50, 168, 0.2)",
                borderColor: "rgba(168, 50, 168, 1)",
                data: mem_v2_sd_full_nullvm_both_master,
                hidden: true,
                fill: false
            }
        ]
    },
    options: memOptions
});