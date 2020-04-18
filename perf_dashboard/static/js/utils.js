function convertData(data) {
  var newData = {};

  newData.datasets = data.datasets.map((dataset) => {
    return {
      label: dataset.label,
      backgroundColor: dataset.backgroundColor,
      borderColor: dataset.borderColor,
      hidden: dataset.hidden,
      fill: dataset.fill,
      data:
        dataset.data.map((d, i) => {
          return {
          x: data.labels[i],
          y: d,
        }
      })
    };
  });

  return newData;
}
