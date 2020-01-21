/* Latency Charts */

// labels along the x-axis
var connNum = ["10", "100", "500", "1000", "2000", "3000"];
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
                labelString: "QPS"
            }
        }]
    }
};

window.onload = function () {
    generateLatencyChart(connNum, options)
};