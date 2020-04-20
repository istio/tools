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
            type: "linear",
            scaleLabel: {
                display: true,
                labelString: "Connections"
            }
        }]
    }
};

window.onload = function () {
    generateLatencyChart(connNum, options)
};

