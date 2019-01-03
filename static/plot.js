var metricsUrl = 'http://127.0.0.1:8080/metrics';

if (window.location.href.includes("com")) {
  metricsUrl = 'https://home-metrics.appspot.com/metrics'
}

Plotly.d3.json(metricsUrl, function(error, response) {
  if (error) {
    return console.error(error);
  }

  var x = [];
  var y = [];
  var events = response.Events;

  if (!events) {
    return console.log("No events to plot.");
  }
  
  for (i = 0; i < events.length; i++) { 
    x.push(new Date(events[i].published_at));
    y.push(parseFloat(events[i].data));
  }

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
