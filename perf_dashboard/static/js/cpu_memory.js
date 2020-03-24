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
                data: cpu_mixer_base,
                fill: false
            }, {
                label: "mixer-both",
                backgroundColor: "rgba(66, 133, 246, 0.2)",
                borderColor: "rgba(66, 133, 246, 1)",
                data: cpu_mixer_both,
                fill: false
            }, {
                label: "none-mtls-both",
                backgroundColor: "rgba(0, 0, 0, 0.2)",
                borderColor: "rgba(0, 0, 0, 1)",
                data: cpu_none_both,
                fill: false
            },{
                label: "none-plaintext-both",
                backgroundColor: "rgba(52, 235, 219, 0.2)",
                borderColor: "rgba(52, 235, 219, 1)",
                data: cpu_none_plaintext_both,
                fill: false
            }, {
                label: "v2-both",
                backgroundColor: "rgba(252, 123, 3, 0.2)",
                borderColor: "rgba(252, 123, 3, 1)",
                data: cpu_v2_both,
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
                data: mem_mixer_base,
                fill: false
            },  {
                label: "mixer-both",
                backgroundColor: "rgba(66, 133, 246, 0.2)",
                borderColor: "rgba(66, 133, 246, 1)",
                data: mem_mixer_both,
                fill: false
            }, {
                label: "none-mtls-both",
                backgroundColor: "rgba(0, 0, 0, 0.2)",
                borderColor: "rgba(0, 0, 0, 1)",
                data: mem_none_both,
                fill: false
            }, {
                label: "none-plaintext-both",
                backgroundColor: "rgba(52, 235, 219, 0.2)",
                borderColor: "rgba(52, 235, 219, 1)",
                data: mem_none_plaintext_both,
                fill: false
            }, {
                label: "v2-both",
                backgroundColor: "rgba(252, 123, 3, 0.2)",
                borderColor: "rgba(252, 123, 3, 1)",
                data: mem_v2_both,
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
                data: cpu_mixer_base_master,
                fill: false
            }, {
                label: "mixer-both",
                backgroundColor: "rgba(66, 133, 246, 0.2)",
                borderColor: "rgba(66, 133, 246, 1)",
                data: cpu_mixer_both_master,
                fill: false
            },  {
                label: "none-mtls-both",
                backgroundColor: "rgba(0, 0, 0, 0.2)",
                borderColor: "rgba(0, 0, 0, 1)",
                data: cpu_none_both_master,
                fill: false
            }, {
                label: "none-plaintext-both",
                backgroundColor: "rgba(52, 235, 219, 0.2)",
                borderColor: "rgba(52, 235, 219, 1)",
                data: cpu_none_plaintext_both_master,
                fill: false
            }, {
                label: "v2-both",
                backgroundColor: "rgba(252, 123, 3, 0.2)",
                borderColor: "rgba(252, 123, 3, 1)",
                data: cpu_v2_both_master,
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
                data: mem_mixer_base_master,
                fill: false
            },  {
                label: "mixer-both",
                backgroundColor: "rgba(66, 133, 246, 0.2)",
                borderColor: "rgba(66, 133, 246, 1)",
                data: mem_mixer_both_master,
                fill: false
            }, {
                label: "none-mtls-both",
                backgroundColor: "rgba(0, 0, 0, 0.2)",
                borderColor: "rgba(0, 0, 0, 1)",
                data: mem_none_both_master,
                fill: false
            }, {
                label: "none-plaintext-both",
                backgroundColor: "rgba(52, 235, 219, 0.2)",
                borderColor: "rgba(52, 235, 219, 1)",
                data: mem_none_plaintext_both_master,
                fill: false
            }, {
                label: "v2-both",
                backgroundColor: "rgba(252, 123, 3, 0.2)",
                borderColor: "rgba(252, 123, 3, 1)",
                data: mem_v2_both_master,
                fill: false
            }
        ]
    },
    options: memOptions
});