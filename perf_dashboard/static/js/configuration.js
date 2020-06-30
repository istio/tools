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