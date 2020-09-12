/* Latency Charts */
// labels along the x-axis
var qpsNum = ["10", "100", "200", "400", "800", "1000"];
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
            type: "linear",
            scaleLabel: {
                display: true,
                labelString: "QPS"
            }
        }]
    }
};

window.onload = function () {
    generateLatencyChart(qpsNum, options)
};
