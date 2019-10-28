/*global $, document, LINECHARTEXMPLE*/
$(document).ready(function () {

    'use strict';

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

    // CPU dataset for drawing the line
    var baselineCPU = [0, 0, 0, 0, 0, 0];
    var serveronlyCPU = [0, 70, 210, 400, 750, 1020];
    var bothCPU = [0, 75, 230, 480, 1000, 1200];
    var nomixerServeronlyCPU = [0, 30, 170, 290, 670, 900];
    var nomixerBothCPU = [0, 65, 235, 565, 820, 1400];

    // Memory dataset for drawing the line
    var baselineMem = [0, 0, 0, 0, 0, 0];
    var serveronlyMem = [0, 70, 210, 400, 750, 1020];
    var bothMem = [0, 75, 230, 480, 1000, 1200];
    var nomixerServeronlyMem = [0, 30, 170, 290, 670, 900];
    var nomixerBothMem = [0, 65, 235, 565, 820, 1400];

    new Chart(document.getElementById("cpu-qps"), {
        type: 'line',
        data: {
            labels: qpsNum,
            datasets: [
                {
                    label: "baseline",
                    backgroundColor: "rgba(236, 66, 53,0.2)",
                    borderColor: "rgba(236, 66, 53,1)",
                    data: baselineCPU,
                    fill: false
                }, {
                    label: "serveronly",
                    backgroundColor: "rgba(259, 188, 5,0.2)",
                    borderColor: "rgba(259, 188, 5,1)",
                    data: serveronlyCPU,
                    fill: false
                }, {
                    label: "both",
                    backgroundColor: "rgba(66, 133, 246,0.2)",
                    borderColor: "rgba(66, 133, 246, 1)",
                    data: bothCPU,
                    fill: false
                }, {
                    label: "nomixer-serveronly",
                    backgroundColor: "rgba(52, 168, 85,0.2)",
                    borderColor: "rgba(52, 168, 85,1)",
                    data: nomixerServeronlyCPU,
                    fill: false
                }, {
                    label: "nomixer-both",
                    backgroundColor: "rgba(0,0,0,0.2)",
                    borderColor: "rgba(0,0,0,1)",
                    data: nomixerBothCPU,
                    fill: false
                }
            ]
        },
        options: cpuOptions
    });

    new Chart(document.getElementById("mem-qps"), {
        type: 'line',
        data: {
            labels: qpsNum,
            datasets: [
                {
                    label: "baseline",
                    backgroundColor: "rgba(236, 66, 53,0.2)",
                    borderColor: "rgba(236, 66, 53,1)",
                    data: baselineMem,
                    fill: false
                }, {
                    label: "serveronly",
                    backgroundColor: "rgba(259, 188, 5,0.2)",
                    borderColor: "rgba(259, 188, 5,1)",
                    data: serveronlyMem,
                    fill: false
                }, {
                    label: "both",
                    backgroundColor: "rgba(66, 133, 246,0.2)",
                    borderColor: "rgba(66, 133, 246, 1)",
                    data: bothMem,
                    fill: false
                }, {
                    label: "nomixer-serveronly",
                    backgroundColor: "rgba(52, 168, 85,0.2)",
                    borderColor: "rgba(52, 168, 85,1)",
                    data: nomixerServeronlyMem,
                    fill: false
                }, {
                    label: "nomixer-both",
                    backgroundColor: "rgba(0,0,0,0.2)",
                    borderColor: "rgba(0,0,0,1)",
                    data: nomixerBothMem,
                    fill: false
                }
            ]
        },
        options: memOptions
    });

});
