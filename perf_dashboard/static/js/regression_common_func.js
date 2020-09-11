window.constructDataSeries = function(trending_data, trending_data_baseline) {
  chart_trending_data = [];
  for (i = 0; i < trending_data.length; i++) {
    if(trending_data[i][3] === 'null') {
      chart_trending_data.push({ x: new Date(trending_data[i][0],
              trending_data[i][1]-1, trending_data[i][2]),
          y: null })
    }
    else {
        if(trending_data[i][3]-trending_data_baseline[i][3] < 0) {
            y_data = null
        } else {
            y_data = trending_data[i][3]-trending_data_baseline[i][3]
        }
        chart_trending_data.push({ x: new Date(trending_data[i][0],
              trending_data[i][1]-1, trending_data[i][2]), y: y_data})}
  }
  return chart_trending_data
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
      title: "Latency trending in milliseconds",
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
            {
            type: "line",
            showInLegend: true,
            name: "none-mtls_both - baseline",
            markerType: "square",
            xValueFormatString: "DD MMM, YYYY",
            color: "rgba(0, 0, 0, 1)",
            dataPoints: modesData[0]
            },
            {
            type: "line",
            showInLegend: true,
            name: "none-plaintext_both - baseline",
            markerType: "square",
            xValueFormatString: "DD MMM, YYYY",
            color: "rgba(52, 235, 219, 1)",
            dataPoints: modesData[1]
            },
            {
            type: "line",
            showInLegend: true,
            name: "v2-stats-nullvm_both - baseline",
            markerType: "square",
            xValueFormatString: "DD MMM, YYYY",
            color: "rgba(252, 123, 3, 1)",
            dataPoints: modesData[2]
            },
            {
            type: "line",
            showInLegend: true,
            name: "v2-stats-wasm_both - baseline",
            markerType: "square",
            xValueFormatString: "DD MMM, YYYY",
            color: "rgba(242, 245, 66, 1)",
            dataPoints: modesData[3]
            },
            {
            type: "line",
            showInLegend: true,
            name: "v2-sd-nologging-nullvm_both - baseline",
            markerType: "square",
            xValueFormatString: "DD MMM, YYYY",
            color: "rgba(52, 168, 85, 1)",
            dataPoints: modesData[4]
            },
            {
            type: "line",
            showInLegend: true,
            name: "v2-sd-full-nullvm_both - baseline",
            markerType: "square",
            xValueFormatString: "DD MMM, YYYY",
            color: "rgba(168, 50, 168, 1)",
            dataPoints: modesData[5]
            }
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
