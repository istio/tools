window.constructDataSeries = function(pattern_data, pattern_data_baseline) {
  chart_pattern_data = [];
  /*
    Example of an array in pattern_data:
    ['2019', '12', '23', 4.478, '20191223', 'release-1.4.20191223-16.da6d4af0a2e1c3207edfd97f09d07a638c59e89a']
   */

  for (i = 0; i < pattern_data.length; i++) {
    perf_num = pattern_data[i][3];
    perf_num_baseline = pattern_data_baseline[i][3];
    year = pattern_data[i][0];
    month = pattern_data[i][1];
    date = pattern_data[i][2];
    if(perf_num === 'null') {
      chart_pattern_data.push({ x: new Date(year,month-1, date), y: null })
    }
    else {
      chart_pattern_data.push({ x: new Date(year,month-1, date), y:perf_num-perf_num_baseline})}
  }
  return chart_pattern_data
};

window.generateChart = function(chartID, modesData) {
  let chart = new CanvasJS.Chart(chartID, {
    animationEnabled: true,
    theme: "light2",
    axisX:{
      valueFormatString: "DD MMM",
      crosshair: {
        enabled: true,
        snapToDataPoint: true
      }
    },
    axisY: {
      title: "Latency Pattern (milliseconds)",
      crosshair: {
        enabled: true
      }
    },
    toolTip:{
      shared:true
    },
    legend:{
      cursor:"pointer",
      verticalAlign: "bottom",
      horizontalAlign: "left",
      dockInsidePlotArea: true,
      itemclick: toogleDataSeries
    },
    data: [{
      type: "line",
      showInLegend: true,
      name: "serveronly-baseline",
      markerType: "square",
      xValueFormatString: "DD MMM, YYYY",
      color: "rgba(259, 188, 5,1)",
      dataPoints: modesData[0]
      },
      {
        type: "line",
        showInLegend: true,
        name: "both-baseline",
        markerType: "square",
        xValueFormatString: "DD MMM, YYYY",
        color: "rgba(66, 133, 246, 1)",
        dataPoints: modesData[1]
      },
      {
        type: "line",
        showInLegend: true,
        name: "none_serveronly-baseline",
        markerType: "square",
        xValueFormatString: "DD MMM, YYYY",
        color: "rgba(52, 168, 85, 1)",
        dataPoints: modesData[2]
      },
      {
        type: "line",
        showInLegend: true,
        name: "none_both-baseline",
        markerType: "square",
        xValueFormatString: "DD MMM, YYYY",
        color: "rgba(0, 0, 0, 1)",
        dataPoints: modesData[3]
      },
      {
        type: "line",
        showInLegend: true,
        name: "v2_serveronly-baseline",
        markerType: "square",
        xValueFormatString: "DD MMM, YYYY",
        color: "rgba(168, 50, 168, 1)",
        dataPoints: modesData[4]
      },
      {
        type: "line",
        showInLegend: true,
        name: "v2_both-baseline",
        markerType: "square",
        xValueFormatString: "DD MMM, YYYY",
        color: "rgba(252, 123, 3, 1)",
        dataPoints: modesData[5]
      },
    ]
  });
  chart.render();
};

toogleDataSeries = function(e, chart) {
    if (typeof (e.dataSeries.visible) === "undefined" || e.dataSeries.visible) {
      e.dataSeries.visible = false;
    } else {
      e.dataSeries.visible = true;
    }
    chart.render();
};
