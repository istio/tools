window.onload = function () {
  let chart_master_pattern_mixer_serveronly = [];
  let chart_master_pattern_mixer_both = [];
  let chart_master_pattern_nomixer_serveronly = [];
  let chart_master_pattern_nomixer_both = [];
  let chart_master_pattern_v2_serveronly = [];
  let chart_master_pattern_v2_both = [];

  for (i = 0; i < master_pattern_mixer_serveronly.length; i++) {
    if(master_pattern_mixer_serveronly[i][3] === "null") {
      chart_master_pattern_mixer_serveronly.push({ x: new Date(master_pattern_mixer_serveronly[i][0],
              master_pattern_mixer_serveronly[i][1]-1, master_pattern_mixer_serveronly[i][2]),
          y: null })
    } else {
      chart_master_pattern_mixer_serveronly.push({ x: new Date(master_pattern_mixer_serveronly[i][0],
              master_pattern_mixer_serveronly[i][1]-1, master_pattern_mixer_serveronly[i][2]),
          y: master_pattern_mixer_serveronly[i][3]-master_pattern_mixer_base[i][3]})}
  }

  for (i = 0; i < master_pattern_mixer_both.length; i++) {
    if(master_pattern_mixer_both[i][3] === "null") {
      chart_master_pattern_mixer_both.push({ x: new Date(master_pattern_mixer_both[i][0],
              master_pattern_mixer_both[i][1]-1, master_pattern_mixer_both[i][2]),
          y: null })
    } else {
      chart_master_pattern_mixer_both.push({ x: new Date(master_pattern_mixer_both[i][0],
              master_pattern_mixer_both[i][1]-1, master_pattern_mixer_both[i][2]),
          y: master_pattern_mixer_both[i][3]-master_pattern_mixer_base[i][3]})}
  }

  for (i = 0; i < master_pattern_nomixer_serveronly.length; i++) {
    if(master_pattern_nomixer_serveronly[i][3] === "null") {
      chart_master_pattern_nomixer_serveronly.push({ x: new Date(master_pattern_nomixer_serveronly[i][0],
              master_pattern_nomixer_serveronly[i][1]-1, master_pattern_nomixer_serveronly[i][2]),
          y: null })
    } else {
      chart_master_pattern_nomixer_serveronly.push({ x: new Date(master_pattern_nomixer_serveronly[i][0],
              master_pattern_nomixer_serveronly[i][1]-1, master_pattern_nomixer_serveronly[i][2]),
          y: master_pattern_nomixer_serveronly[i][3]-master_pattern_mixer_base[i][3]})}
  }

  for (i = 0; i < master_pattern_nomixer_both.length; i++) {
    if(master_pattern_nomixer_both[i][3] === "null") {
      chart_master_pattern_nomixer_both.push({ x: new Date(master_pattern_nomixer_both[i][0],
              master_pattern_nomixer_both[i][1]-1, master_pattern_nomixer_both[i][2]),
          y: null })
    } else {
      chart_master_pattern_nomixer_both.push({ x: new Date(master_pattern_nomixer_both[i][0],
              master_pattern_nomixer_both[i][1]-1, master_pattern_nomixer_both[i][2]),
          y: master_pattern_nomixer_both[i][3]-master_pattern_mixer_base[i][3]})}
  }

  for (i = 0; i < master_pattern_v2_serveronly.length; i++) {
    if(master_pattern_v2_serveronly[i][3] === 'null') {
      chart_master_pattern_v2_serveronly.push({ x: new Date(master_pattern_v2_serveronly[i][0],
              master_pattern_v2_serveronly[i][1]-1, master_pattern_v2_serveronly[i][2]),
          y: null })
    }
    else {
      chart_master_pattern_v2_serveronly.push({ x: new Date(master_pattern_v2_serveronly[i][0],
              master_pattern_v2_serveronly[i][1]-1, master_pattern_v2_serveronly[i][2]),
          y:master_pattern_v2_serveronly[i][3]-master_pattern_mixer_base[i][3]})}
  }

  for (i = 0; i < master_pattern_v2_both.length; i++) {
    if(master_pattern_v2_both[i][3] === "null") {
      chart_master_pattern_v2_both.push({ x: new Date(master_pattern_v2_both[i][0],
              master_pattern_v2_both[i][1]-1, master_pattern_v2_both[i][2]),
          y: null })
    } else {
      chart_master_pattern_v2_both.push({ x: new Date(master_pattern_v2_both[i][0],
              master_pattern_v2_both[i][1]-1, master_pattern_v2_both[i][2]),
          y:master_pattern_v2_both[i][3]-master_pattern_mixer_base[i][3] })}
  }

  let chart = new CanvasJS.Chart("chartContainer", {
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
      dataPoints: chart_master_pattern_mixer_serveronly},
      {
        type: "line",
        showInLegend: true,
        name: "both-baseline",
        markerType: "square",
        xValueFormatString: "DD MMM, YYYY",
        color: "rgba(66, 133, 246, 1)",
        dataPoints: chart_master_pattern_mixer_both},
      {
        type: "line",
        showInLegend: true,
        name: "nomixer_serveronly-baseline",
        markerType: "square",
        xValueFormatString: "DD MMM, YYYY",
        color: "rgba(52, 168, 85,1)",
        dataPoints: chart_master_pattern_nomixer_serveronly},
      {
        type: "line",
        showInLegend: true,
        name: "nomixer_both-baseline",
        markerType: "square",
        xValueFormatString: "DD MMM, YYYY",
        color: "rgba(0, 0, 0, 1)",
        dataPoints: chart_master_pattern_nomixer_both},
      {
        type: "line",
        showInLegend: true,
        name: "v2_serveronly-baseline",
        markerType: "square",
        xValueFormatString: "DD MMM, YYYY",
        color: "rgba(168, 50, 168, 1)",
        dataPoints: chart_master_pattern_v2_serveronly}
        ,{
        type: "line",
        showInLegend: true,
        name: "v2_both-baseline",
        markerType: "square",
        xValueFormatString: "DD MMM, YYYY",
        color: "rgba(252, 123, 3, 1)",
        dataPoints: chart_master_pattern_v2_both},
    ]
  });
  chart.render();

  function toogleDataSeries(e){
    if (typeof(e.dataSeries.visible) === "undefined" || e.dataSeries.visible) {
      e.dataSeries.visible = false;
    } else{
      e.dataSeries.visible = true;
    }
    chart.render();
  }
}
