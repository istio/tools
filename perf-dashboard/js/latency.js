/*global $, document, LINECHARTEXMPLE*/
$(document).ready(function () {

    'use strict';

    // labels along the x-axis
    var connNum = ["1", "2", "4", "8", "16", "32", "64"];
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

    // P50 dataset for drawing the line
    var baselineP50 = [0.380, 0.480, 0.550, 0.630, 0.820, 1.890, 3.040];
    var serveronlyP50 = [0.820, 0.970, 1.250, 1.730, 3.100, 5.960, 10.590];
    var bothP50 = [1.480, 1.770, 2.220, 3.190, 7.970, 13.700, 21.500];
    var nomixerServeronlyP50 = [0.610, 0.780, 1.050, 1.370, 1.910, 3.030, 5.200];
    var nomixerBothP50 = [0.990, 1.240, 1.650, 2.180, 3.340, 5.320, 8.990];

    // P90 dataset for drawing the line
    var baselineP90 = [0.380, 0.480, 0.550, 0.630, 0.820, 1.890, 3.040];
    var serveronlyP90 = [0.820, 0.970, 1.250, 1.730, 3.100, 5.960, 10.590];
    var bothP90 = [1.480, 1.770, 2.220, 3.190, 7.970, 13.700, 21.500];
    var nomixerServeronlyP90 = [0.610, 0.780, 1.050, 1.370, 1.910, 3.030, 5.200];
    var nomixerBothP90 = [0.990, 1.240, 1.650, 2.180, 3.340, 5.320, 8.990];

    // P99 dataset for drawing the line
    var baselineP99 = [0.380, 0.480, 0.550, 0.630, 0.820, 1.890, 3.040];
    var serveronlyP99 = [0.820, 0.970, 1.250, 1.730, 3.100, 5.960, 10.590];
    var bothP99 = [1.480, 1.770, 2.220, 3.190, 7.970, 13.700, 21.500];
    var nomixerServeronlyP99 = [0.610, 0.780, 1.050, 1.370, 1.910, 3.030, 5.200];
    var nomixerBothP99 = [0.990, 1.240, 1.650, 2.180, 3.340, 5.320, 8.990];

    new Chart(document.getElementById("latency-p50"), {
        type: 'line',
        data: {
            labels: connNum,
            datasets: [
                {
                    label: "baseline",
                    backgroundColor: "rgba(236, 66, 53,0.2)",
                    borderColor: "rgba(236, 66, 53,1)",
                    data: baselineP50,
                    fill: false
                }, {
                    label: "serveronly",
                    backgroundColor: "rgba(259, 188, 5,0.2)",
                    borderColor: "rgba(259, 188, 5,1)",
                    data: serveronlyP50,
                    fill: false
                }, {
                    label: "both",
                    backgroundColor: "rgba(66, 133, 246,0.2)",
                    borderColor: "rgba(66, 133, 246, 1)",
                    data: bothP50,
                    fill: false
                }, {
                    label: "nomixer-serveronly",
                    backgroundColor: "rgba(52, 168, 85,0.2)",
                    borderColor: "rgba(52, 168, 85,1)",
                    data: nomixerServeronlyP50,
                    fill: false
                }, {
                    label: "nomixer-both",
                    backgroundColor: "rgba(0,0,0,0.2)",
                    borderColor: "rgba(0,0,0,1)",
                    data: nomixerBothP50,
                    fill: false
                }
            ]
        },
        options: options
    });

    new Chart(document.getElementById("latency-p90"), {
        type: 'line',
        data: {
            labels: connNum,
            datasets: [
                {
                    label: "baseline",
                    backgroundColor: "rgba(236, 66, 53,0.2)",
                    borderColor: "rgba(236, 66, 53,1)",
                    data: baselineP90,
                    fill: false
                }, {
                    label: "serveronly",
                    backgroundColor: "rgba(259, 188, 5,0.2)",
                    borderColor: "rgba(259, 188, 5,1)",
                    data: serveronlyP90,
                    fill: false
                }, {
                    label: "both",
                    backgroundColor: "rgba(66, 133, 246,0.2)",
                    borderColor: "rgba(66, 133, 246, 1)",
                    data: bothP90,
                    fill: false
                }, {
                    label: "nomixer-serveronly",
                    backgroundColor: "rgba(52, 168, 85,0.2)",
                    borderColor: "rgba(52, 168, 85,1)",
                    data: nomixerServeronlyP90,
                    fill: false
                }, {
                    label: "nomixer-both",
                    backgroundColor: "rgba(0,0,0,0.2)",
                    borderColor: "rgba(0,0,0,1)",
                    data: nomixerBothP90,
                    fill: false
                }
            ]
        },
        options: options
    });


    new Chart(document.getElementById("latency-p99"), {
        type: 'line',
        data: {
            labels: connNum,
            datasets: [
                {
                    label: "baseline",
                    backgroundColor: "rgba(236, 66, 53,0.2)",
                    borderColor: "rgba(236, 66, 53,1)",
                    data: baselineP99,
                    fill: false
                }, {
                    label: "serveronly",
                    backgroundColor: "rgba(259, 188, 5,0.2)",
                    borderColor: "rgba(259, 188, 5,1)",
                    data: serveronlyP99,
                    fill: false
                }, {
                    label: "both",
                    backgroundColor: "rgba(66, 133, 246,0.2)",
                    borderColor: "rgba(66, 133, 246, 1)",
                    data: bothP99,
                    fill: false
                }, {
                    label: "nomixer-serveronly",
                    backgroundColor: "rgba(52, 168, 85,0.2)",
                    borderColor: "rgba(52, 168, 85,1)",
                    data: nomixerServeronlyP99,
                    fill: false
                }, {
                    label: "nomixer-both",
                    backgroundColor: "rgba(0,0,0,0.2)",
                    borderColor: "rgba(0,0,0,1)",
                    data: nomixerBothP99,
                    fill: false
                }
            ]
        },
        options: options
    });
});
