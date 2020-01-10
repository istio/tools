/* Latency Charts */

// labels along the x-axis
var connNum = ["2", "4", "8", "16", "32", "64"];
// x-axis and y-axis labels
var options = {
    scales: {
        yAxes: [{
            scaleLabel: {
                display: true,
                labelString: "Latency in milliseconds"
            }
        }],
        xAxes: [{
            scaleLabel: {
                display: true,
                labelString: "Connections"
            }
        }]
    }
};


new Chart(document.getElementById('latency-config-demo'), {
    type: 'line',
    data: {
        labels: connNum,
        datasets: [
            {
                label: "baseline",
                backgroundColor: "rgba(236, 66, 53,0.2)",
                borderColor: "rgba(236, 66, 53,1)",
                data: latency_mixer_base_p90_master,
                fill: false
            }, {
                label: "serveronly",
                backgroundColor: "rgba(259, 188, 5,0.2)",
                borderColor: "rgba(259, 188, 5,1)",
                data: latency_mixer_serveronly_p90_master,
                fill: false
            }, {
                label: "both",
                backgroundColor: "rgba(66, 133, 246,0.2)",
                borderColor: "rgba(66, 133, 246, 1)",
                data: latency_mixer_both_p90_master,
                fill: false
            }, {
                label: "none-serveronly",
                backgroundColor: "rgba(52, 168, 85,0.2)",
                borderColor: "rgba(52, 168, 85,1)",
                data: latency_none_serveronly_p90_master,
                fill: false
            }, {
                label: "none-both",
                backgroundColor: "rgba(0,0,0,0.2)",
                borderColor: "rgba(0,0,0,1)",
                data: latency_none_both_p90_master,
                fill: false
            }, {
                label: "v2-serveronly",
                backgroundColor: "rgba(168, 50, 168, 0.2)",
                borderColor: "rgba(168, 50, 168, 1)",
                data: latency_v2_serveronly_p90_master,
                fill: false
            }, {
                label: "v2-both",
                backgroundColor: "rgba(252, 123, 3, 0.2)",
                borderColor: "rgba(252, 123, 3, 1)",
                data: latency_v2_both_p90_master,
                fill: false
            }
        ]
    },
    options: options
});