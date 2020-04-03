window.constructDataSeries = function(pattern_data, pattern_data_baseline) {
  chart_pattern_data = [];
  for (i = 0; i < pattern_data.length; i++) {
    if(pattern_data[i][3] === 'null') {
      chart_pattern_data.push({ x: new Date(pattern_data[i][0],
              pattern_data[i][1]-1, pattern_data[i][2]),
          y: null })
    }
    else {
      chart_pattern_data.push({ x: new Date(pattern_data[i][0],
              pattern_data[i][1]-1, pattern_data[i][2]),
          y:pattern_data[i][3]-pattern_data_baseline[i][3]})}
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
      title: "p90 Latency Pattern in milliseconds",
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
      itemclick: toggleDataSeries
    },
    data: [
      //   {
      //   type: "line",
      //   showInLegend: true,
      //   name: "baseline",
      //   markerType: "square",
      //   xValueFormatString: "DD MMM, YYYY",
      //   color: "rgba(236, 66, 53, 1)",
      //   dataPoints: modesData[0]
      // },
      {
        type: "line",
        showInLegend: true,
        name: "mixer_both - baseline",
        markerType: "square",
        xValueFormatString: "DD MMM, YYYY",
        color: "rgba(66, 133, 246, 1)",
        dataPoints: modesData[1]
      }, {
        type: "line",
        showInLegend: true,
        name: "none-mtls_both - baseline",
        markerType: "square",
        xValueFormatString: "DD MMM, YYYY",
        color: "rgba(0, 0, 0, 1)",
        dataPoints: modesData[2]
      },
      {
        type: "line",
        showInLegend: true,
        name: "v2-sd-full-nullvm_both - baseline",
        markerType: "square",
        xValueFormatString: "DD MMM, YYYY",
        color: "rgba(168, 50, 168, 1)",
        dataPoints: modesData[3]
      },
    ]
  });
  chart.render();
};

toggleDataSeries = function(e, chart) {
    if (typeof (e.dataSeries.visible) === "undefined" || e.dataSeries.visible) {
      e.dataSeries.visible = false;
    } else {
      e.dataSeries.visible = true;
    }
    chart.render();
};
