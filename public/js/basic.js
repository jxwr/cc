$.get('/app/info', function(data){

var AppConfig = data.body.AppConfig;
var Leader = data.body.Leader;

var HTTP_HOST = 'http://'+Leader.Ip+':'+Leader.HttpPort;
var WS_HOST = 'ws://'+Leader.Ip+':'+Leader.WsPort;

var GlobalNodes = [];

if (location.origin != HTTP_HOST)
  location = HTTP_HOST + location.pathname;

var openingObserver = Rx.Observer.create(function() { console.log('Opening socket'); });
var closingObserver = Rx.Observer.create(function() { console.log('Closing socket'); });

var AppInfo = React.createClass({
  render: function() {
    var info = this.props.info;
    var appConfig = _.map(info.AppConfig, function(v, k){
      console.log(k,v);
      return <tr><td>{k}</td><td>{v.toString()}</td></tr>;
    });
    var leaderConfig = _.map(info.Leader, function(v, k){
      return <tr><td>{k}</td><td>{v}</td></tr>;
    });
    return (
      <table className="ui table">
        <tbody>
        <tr>
          <th>ZkConfig</th>
          <th>Value</th>
        </tr>
        {appConfig}
        {leaderConfig}
        </tbody>
      </table>
    );
  }
});

React.render(
    <AppInfo info={data.body} />,
    document.getElementById('content')
);

});
