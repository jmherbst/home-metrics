function plotMetrics(limit) {
  var metricsUrl = 'http://127.0.0.1:8080/metrics';
  var queryParams = "";

  if (window.location.href.includes("com")) {
    metricsUrl = 'https://home-metrics.appspot.com/metrics';
  }
  
  if (limit) { // Query default subset of Metrics
    queryParams = '?limit=' + limit;
  } else { // Going to query based off time submitted in calendar
    queryParams = '?time=' + (new Date($('#metrics-calendar').calendar('get date'))).toIsoString();
  }

  Plotly.d3.json(metricsUrl + queryParams, function (error, response) {
    if (error) {
      console.error(error);
      return null;
    }

    metrics = response.Events;
    if (!metrics) {
      return console.log("No metrics to plot. " + metrics);
    }

    var x = [];
    var y = [];
    for (i = 0; i < metrics.length; i++) {
      x.push(new Date(metrics[i].published_at));
      y.push(parseFloat(metrics[i].data));
    }

    console.log("First date out of " + x.length + ": " + x[0]);
    console.log("Last date out of " + x.length + ": " + x[x.length-1]);
    var graphData = [{
      x: x,
      y: y,
      mode: 'lines+markers',
      text: x,
      name: 'inches',
      type: 'scatter'
    }]

    layout = {
      title: 'Depth of Water',
      xaxis: {},
      yaxis: {
        title: 'Water Depth (inches)',
        titlefont: {
          family: 'Courier New, monospace',
          size: 18,
          color: '#7f7f7f'
        }
      }
    }

    Plotly.newPlot('SumpPumpPlot', graphData, layout, { showSendToCloud: false });
  });
}
