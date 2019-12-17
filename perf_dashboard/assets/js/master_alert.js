window.onload = function () {
   // p90
  let chart_master_pattern_mixer_serveronly_p90 = [];
  let chart_master_pattern_mixer_both_p90 = [];
  let chart_master_pattern_nomixer_serveronly_p90 = [];
  let chart_master_pattern_nomixer_both_p90 = [];
  let chart_master_pattern_v2_serveronly_p90 = [];
  let chart_master_pattern_v2_both_p90 = [];

  for (i = 0; i < master_pattern_mixer_serveronly_p90.length; i++) {
    if(master_pattern_mixer_serveronly_p90[i][3] === "null") {
      chart_master_pattern_mixer_serveronly_p90.push({ x: new Date(master_pattern_mixer_serveronly_p90[i][0],
              master_pattern_mixer_serveronly_p90[i][1]-1, master_pattern_mixer_serveronly_p90[i][2]),
          y: null })
    } else {
      chart_master_pattern_mixer_serveronly_p90.push({ x: new Date(master_pattern_mixer_serveronly_p90[i][0],
              master_pattern_mixer_serveronly_p90[i][1]-1, master_pattern_mixer_serveronly_p90[i][2]),
          y: master_pattern_mixer_serveronly_p90[i][3]-master_pattern_mixer_base_p90[i][3]})}
  }

  for (i = 0; i < master_pattern_mixer_both_p90.length; i++) {
    if(master_pattern_mixer_both_p90[i][3] === "null") {
      chart_master_pattern_mixer_both_p90.push({ x: new Date(master_pattern_mixer_both_p90[i][0],
              master_pattern_mixer_both_p90[i][1]-1, master_pattern_mixer_both_p90[i][2]),
          y: null })
    } else {
      chart_master_pattern_mixer_both_p90.push({ x: new Date(master_pattern_mixer_both_p90[i][0],
              master_pattern_mixer_both_p90[i][1]-1, master_pattern_mixer_both_p90[i][2]),
          y: master_pattern_mixer_both_p90[i][3]-master_pattern_mixer_base_p90[i][3]})}
  }

  for (i = 0; i < master_pattern_nomixer_serveronly_p90.length; i++) {
    if(master_pattern_nomixer_serveronly_p90[i][3] === "null") {
      chart_master_pattern_nomixer_serveronly_p90.push({ x: new Date(master_pattern_nomixer_serveronly_p90[i][0],
              master_pattern_nomixer_serveronly_p90[i][1]-1, master_pattern_nomixer_serveronly_p90[i][2]),
          y: null })
    } else {
      chart_master_pattern_nomixer_serveronly_p90.push({ x: new Date(master_pattern_nomixer_serveronly_p90[i][0],
              master_pattern_nomixer_serveronly_p90[i][1]-1, master_pattern_nomixer_serveronly_p90[i][2]),
          y: master_pattern_nomixer_serveronly_p90[i][3]-master_pattern_mixer_base_p90[i][3]})}
  }

  for (i = 0; i < master_pattern_nomixer_both_p90.length; i++) {
    if(master_pattern_nomixer_both_p90[i][3] === "null") {
      chart_master_pattern_nomixer_both_p90.push({ x: new Date(master_pattern_nomixer_both_p90[i][0],
              master_pattern_nomixer_both_p90[i][1]-1, master_pattern_nomixer_both_p90[i][2]),
          y: null })
    } else {
      chart_master_pattern_nomixer_both_p90.push({ x: new Date(master_pattern_nomixer_both_p90[i][0],
              master_pattern_nomixer_both_p90[i][1]-1, master_pattern_nomixer_both_p90[i][2]),
          y: master_pattern_nomixer_both_p90[i][3]-master_pattern_mixer_base_p90[i][3]})}
  }

  for (i = 0; i < master_pattern_v2_serveronly_p90.length; i++) {
    if(master_pattern_v2_serveronly_p90[i][3] === 'null') {
      chart_master_pattern_v2_serveronly_p90.push({ x: new Date(master_pattern_v2_serveronly_p90[i][0],
              master_pattern_v2_serveronly_p90[i][1]-1, master_pattern_v2_serveronly_p90[i][2]),
          y: null })
    }
    else {
      chart_master_pattern_v2_serveronly_p90.push({ x: new Date(master_pattern_v2_serveronly_p90[i][0],
              master_pattern_v2_serveronly_p90[i][1]-1, master_pattern_v2_serveronly_p90[i][2]),
          y:master_pattern_v2_serveronly_p90[i][3]-master_pattern_mixer_base_p90[i][3]})}
  }

  for (i = 0; i < master_pattern_v2_both_p90.length; i++) {
    if(master_pattern_v2_both_p90[i][3] === "null") {
      chart_master_pattern_v2_both_p90.push({ x: new Date(master_pattern_v2_both_p90[i][0],
              master_pattern_v2_both_p90[i][1]-1, master_pattern_v2_both_p90[i][2]),
          y: null })
    } else {
      chart_master_pattern_v2_both_p90.push({ x: new Date(master_pattern_v2_both_p90[i][0],
              master_pattern_v2_both_p90[i][1]-1, master_pattern_v2_both_p90[i][2]),
          y:master_pattern_v2_both_p90[i][3]-master_pattern_mixer_base_p90[i][3] })}
  }

  // p99
  let chart_master_pattern_mixer_serveronly_p99 = [];
  let chart_master_pattern_mixer_both_p99 = [];
  let chart_master_pattern_nomixer_serveronly_p99 = [];
  let chart_master_pattern_nomixer_both_p99 = [];
  let chart_master_pattern_v2_serveronly_p99 = [];
  let chart_master_pattern_v2_both_p99 = [];

  for (i = 0; i < master_pattern_mixer_serveronly_p99.length; i++) {
    if(master_pattern_mixer_serveronly_p99[i][3] === "null") {
      chart_master_pattern_mixer_serveronly_p99.push({ x: new Date(master_pattern_mixer_serveronly_p99[i][0],
              master_pattern_mixer_serveronly_p99[i][1]-1, master_pattern_mixer_serveronly_p99[i][2]),
          y: null })
    } else {
      chart_master_pattern_mixer_serveronly_p99.push({ x: new Date(master_pattern_mixer_serveronly_p99[i][0],
              master_pattern_mixer_serveronly_p99[i][1]-1, master_pattern_mixer_serveronly_p99[i][2]),
          y: master_pattern_mixer_serveronly_p99[i][3]-master_pattern_mixer_base_p99[i][3]})}
  }

  for (i = 0; i < master_pattern_mixer_both_p99.length; i++) {
    if(master_pattern_mixer_both_p99[i][3] === "null") {
      chart_master_pattern_mixer_both_p99.push({ x: new Date(master_pattern_mixer_both_p99[i][0],
              master_pattern_mixer_both_p99[i][1]-1, master_pattern_mixer_both_p99[i][2]),
          y: null })
    } else {
      chart_master_pattern_mixer_both_p99.push({ x: new Date(master_pattern_mixer_both_p99[i][0],
              master_pattern_mixer_both_p99[i][1]-1, master_pattern_mixer_both_p99[i][2]),
          y: master_pattern_mixer_both_p99[i][3]-master_pattern_mixer_base_p99[i][3]})}
  }

  for (i = 0; i < master_pattern_nomixer_serveronly_p99.length; i++) {
    if(master_pattern_nomixer_serveronly_p99[i][3] === "null") {
      chart_master_pattern_nomixer_serveronly_p99.push({ x: new Date(master_pattern_nomixer_serveronly_p99[i][0],
              master_pattern_nomixer_serveronly_p99[i][1]-1, master_pattern_nomixer_serveronly_p99[i][2]),
          y: null })
    } else {
      chart_master_pattern_nomixer_serveronly_p99.push({ x: new Date(master_pattern_nomixer_serveronly_p99[i][0],
              master_pattern_nomixer_serveronly_p99[i][1]-1, master_pattern_nomixer_serveronly_p99[i][2]),
          y: master_pattern_nomixer_serveronly_p99[i][3]-master_pattern_mixer_base_p99[i][3]})}
  }

  for (i = 0; i < master_pattern_nomixer_both_p99.length; i++) {
    if(master_pattern_nomixer_both_p99[i][3] === "null") {
      chart_master_pattern_nomixer_both_p99.push({ x: new Date(master_pattern_nomixer_both_p99[i][0],
              master_pattern_nomixer_both_p99[i][1]-1, master_pattern_nomixer_both_p99[i][2]),
          y: null })
    } else {
      chart_master_pattern_nomixer_both_p99.push({ x: new Date(master_pattern_nomixer_both_p99[i][0],
              master_pattern_nomixer_both_p99[i][1]-1, master_pattern_nomixer_both_p99[i][2]),
          y: master_pattern_nomixer_both_p99[i][3]-master_pattern_mixer_base_p99[i][3]})}
  }

  for (i = 0; i < master_pattern_v2_serveronly_p99.length; i++) {
    if(master_pattern_v2_serveronly_p99[i][3] === 'null') {
      chart_master_pattern_v2_serveronly_p99.push({ x: new Date(master_pattern_v2_serveronly_p99[i][0],
              master_pattern_v2_serveronly_p99[i][1]-1, master_pattern_v2_serveronly_p99[i][2]),
          y: null })
    }
    else {
      chart_master_pattern_v2_serveronly_p99.push({ x: new Date(master_pattern_v2_serveronly_p99[i][0],
              master_pattern_v2_serveronly_p99[i][1]-1, master_pattern_v2_serveronly_p99[i][2]),
          y:master_pattern_v2_serveronly_p99[i][3]-master_pattern_mixer_base_p99[i][3]})}
  }

  for (i = 0; i < master_pattern_v2_both_p99.length; i++) {
    if(master_pattern_v2_both_p99[i][3] === "null") {
      chart_master_pattern_v2_both_p99.push({ x: new Date(master_pattern_v2_both_p99[i][0],
              master_pattern_v2_both_p99[i][1]-1, master_pattern_v2_both_p99[i][2]),
          y: null })
    } else {
      chart_master_pattern_v2_both_p99.push({ x: new Date(master_pattern_v2_both_p99[i][0],
              master_pattern_v2_both_p99[i][1]-1, master_pattern_v2_both_p99[i][2]),
          y:master_pattern_v2_both_p99[i][3]-master_pattern_mixer_base_p99[i][3] })}
  }

  let chartP90Master = new CanvasJS.Chart("chart_p90_master", {
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
      title: "Latency Pattern in milliseconds",
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
      dataPoints: chart_master_pattern_mixer_serveronly_p90},
      {
        type: "line",
        showInLegend: true,
        name: "both-baseline",
        markerType: "square",
        xValueFormatString: "DD MMM, YYYY",
        color: "rgba(66, 133, 246, 1)",
        dataPoints: chart_master_pattern_mixer_both_p90},
      {
        type: "line",
        showInLegend: true,
        name: "nomixer_serveronly-baseline",
        markerType: "square",
        xValueFormatString: "DD MMM, YYYY",
        color: "rgba(52, 168, 85,1)",
        dataPoints: chart_master_pattern_nomixer_serveronly_p90},
      {
        type: "line",
        showInLegend: true,
        name: "nomixer_both-baseline",
        markerType: "square",
        xValueFormatString: "DD MMM, YYYY",
        color: "rgba(0, 0, 0, 1)",
        dataPoints: chart_master_pattern_nomixer_both_p90},
      {
        type: "line",
        showInLegend: true,
        name: "v2_serveronly-baseline",
        markerType: "square",
        xValueFormatString: "DD MMM, YYYY",
        color: "rgba(168, 50, 168, 1)",
        dataPoints: chart_master_pattern_v2_serveronly_p90}
        ,{
        type: "line",
        showInLegend: true,
        name: "v2_both-baseline",
        markerType: "square",
        xValueFormatString: "DD MMM, YYYY",
        color: "rgba(252, 123, 3, 1)",
        dataPoints: chart_master_pattern_v2_both_p90},
    ]
  });
  chartP90Master.render();

  let chartP99Master = new CanvasJS.Chart("chart_p99_master", {
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
      title: "Latency Pattern in milliseconds",
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
      dataPoints: chart_master_pattern_mixer_serveronly_p99
     },
      {
        type: "line",
        showInLegend: true,
        name: "both-baseline",
        markerType: "square",
        xValueFormatString: "DD MMM, YYYY",
        color: "rgba(66, 133, 246, 1)",
        dataPoints: chart_master_pattern_mixer_both_p99
      },
      {
        type: "line",
        showInLegend: true,
        name: "nomixer_serveronly-baseline",
        markerType: "square",
        xValueFormatString: "DD MMM, YYYY",
        color: "rgba(52, 168, 85,1)",
        dataPoints: chart_master_pattern_nomixer_serveronly_p90
      },
      {
        type: "line",
        showInLegend: true,
        name: "nomixer_both-baseline",
        markerType: "square",
        xValueFormatString: "DD MMM, YYYY",
        color: "rgba(0, 0, 0, 1)",
        dataPoints: chart_master_pattern_nomixer_both_p99
      },
      {
        type: "line",
        showInLegend: true,
        name: "v2_serveronly-baseline",
        markerType: "square",
        xValueFormatString: "DD MMM, YYYY",
        color: "rgba(168, 50, 168, 1)",
        dataPoints: chart_master_pattern_v2_serveronly_p99
      },
      {
        type: "line",
        showInLegend: true,
        name: "v2_both-baseline",
        markerType: "square",
        xValueFormatString: "DD MMM, YYYY",
        color: "rgba(252, 123, 3, 1)",
        dataPoints: chart_master_pattern_v2_both_p99
      },
    ]
  });
  chartP99Master.render();

  function toogleDataSeries(e){
    if (typeof(e.dataSeries.visible) === "undefined" || e.dataSeries.visible) {
      e.dataSeries.visible = false;
    } else{
      e.dataSeries.visible = true;
    }
    chartP90Master.render();
    chartP99Master.render();
  }
}
