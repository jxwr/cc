$.get('/app/info', function(data){

var AppConfig = data.AppConfig;
var Leader = data.Leader;

var HTTP_HOST = 'http://'+Leader.Ip+':'+Leader.HttpPort;
var WS_HOST = 'ws://'+Leader.Ip+':'+Leader.WsPort;

var GlobalNodes = [];

if (location.origin != HTTP_HOST)
  location = HTTP_HOST + location.pathname;

var openingObserver = Rx.Observer.create(function() { console.log('Opening socket'); });
var closingObserver = Rx.Observer.create(function() { console.log('Closing socket'); });

var logSocket = Rx.DOM.fromWebSocket(
  WS_HOST +'/log', null, openingObserver, closingObserver);
var RxLog = logSocket.map(function(e){ 
  var log = JSON.parse(e.data); 
  log.Time = new Date(log.Time).toLocaleString(); 
  return log; 
}).filter(function(e){
  return e.Level != "VERBOSE";
});

var LogPanel = React.createClass({
  getInitialState: function() {
    return {logs: []};
  },
  componentDidMount: function() {
    var self = this;
    RxLog.subscribe(
      function (obj) {
        var logs = self.state.logs;
        if (logs.length > 0 && obj.Message == logs[logs.length-1].Message) {
          var lastLog = logs[logs.length-1];
          // 相同的日志，显示一个计数，避免刷屏
          if (!lastLog.Repeat) lastLog.Repeat = 1;
          lastLog.Time = obj.Time + " ("+ lastLog.Repeat +" times)";
          lastLog.Repeat++;
        } else {
          logs.push(obj);
        }
        logs = logs.slice(-500);
        self.setState({logs: logs});
      },
      function (e) {
        console.log('Error: ', e);
      },
      function (){
        console.log('Closed');
      });
  },
  componentWillUpdate: function() {
    var node = this.getDOMNode();
    this.shouldScrollBottom = node.scrollTop + node.offsetHeight === node.scrollHeight;
  },
  componentDidUpdate: function() {
    if (this.shouldScrollBottom) {
      var node = this.getDOMNode();
      node.scrollTop = node.scrollHeight
    }
  },
  render: function() {
    var logs = this.state.logs;
    var style = {
      maxHeight: "80%",
      overflow: "auto",
    };
    var logRows = _.map(logs, function(obj) {
      return (
          <tr>
            <td>{obj.Level}</td>
            <td>{obj.Time}</td>
            <td>{obj.Target}</td>
            <td>{obj.Message}</td>
          </tr>
        );
    });
    var panel = (
        <div style={style}>
          <table>
            {logRows}
          </table>
        </div>
      );
    return panel;
  }
});

React.render(
    <LogPanel />,
    document.getElementById('content')
);

});
