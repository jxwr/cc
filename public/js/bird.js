$.get('/app/info', function(data){

var AppConfig = data.body.AppConfig;
var Leader = data.body.Leader;

var HTTP_HOST = 'http://'+Leader.Ip+':'+Leader.HttpPort;
var WS_HOST = 'ws://'+Leader.Ip+':'+Leader.WsPort;

var GlobalNodes = [];
var GC = {
  ShowDetail: false
};

if (location.origin != HTTP_HOST)
  location = HTTP_HOST + location.pathname;

var openingObserver = Rx.Observer.create(function() { console.log('Opening socket'); });
var closingObserver = Rx.Observer.create(function() { console.log('Closing socket'); });

var stateSocket = Rx.DOM.fromWebSocket(
  WS_HOST +'/node/state', null, openingObserver, closingObserver);
var RxNodeState = stateSocket.map(function(e){ 
  var state = JSON.parse(e.data); 
  delete state.Room;
  delete state.PFail;
  delete state.ClusterStatsMessagesSent;
  delete state.ClusterStatsMessagesReceived;
  delete state.UsedMemory;
  delete state.Expires;
  delete state.InstantaneousOpsPerSec;
  delete state.InstantaneousInputKbps;
  delete state.InstantaneousOutputKbps;
  return state;
});

var StateColorMap = {
  "RUNNING": "green",
  "OFFLINE": "yellow",
  "WAIT_FAILOVER_BEGIN": "red",
  "WAIT_FAILOVER_END": "red",
};

var RangeBarItem = React.createClass({
  render: function() {
    var width = parseInt(this.props.width);
    var range = this.props.range;
    var style = {
      left: range.Left*width/16384,
      width: (range.Right-range.Left+1)*width/16384,
      backgroundColor: "#5bbd72",
      height: "15px",
      position: "absolute"
    };
    return (
        <div className="rangeBarItem" style={style}>
        </div>
    );
  }
});

var RangeBar = React.createClass({
  render: function() {
    var style = {
      position: "relative",
      width: "800px",
      height: "15px",
    };
    var node = this.props.node;
    var ranges = this.props.ranges;
    var rangePairs = ranges.map(function (range) {
      return [range.Left,range.Right];
    });
    var rangePairTexts = ranges.map(function (range) {
      return range.Left + "-" + range.Right;
    });
    var items = ranges.map(function (range) {
      return (
          <RangeBarItem key={range.Left} range={range} width={style.width} />
      );
    });
    return (
        <td className="rangeBar" style={style}>
          {items}
        </td>
    );
  }
});

var BirdView = React.createClass({
  regionVersion: {},
  lastCheckTime: 0,
  lastRegionShowTime: {},
  getInitialState: function() {
    return {nodes: {}, nodesLastShowTime: {}};
  },
  componentDidMount: function() {
    var self = this;
    RxNodeState.subscribe(
      function (obj) {
        var nodes = self.state.nodes;
        var nodesLastShowTime = self.state.nodesLastShowTime;
        nodes[obj.Id] = obj;
        nodesLastShowTime[obj.Id] = Date.now();
        self.lastRegionShowTime[obj.Region] = Date.now();
        /* <><><> EVIL <><><> */
        GlobalNodes = nodes;
        self.setState({nodes: nodes});
        // 每秒检查一次5s没有汇报的节点删除掉
        var now = Date.now();
        if (now - self.lastCheckTime > 1000) {
          self.lastCheckTime = now;
          for (var key in nodes) {
            var node = nodes[key];
            if (now - self.lastRegionShowTime[obj.Region] < 5000 && now - nodesLastShowTime[node.Id] > 5000) {
              nodesLastShowTime[node.Id] = 0;
              console.log(key, "expired");
              delete nodes[key];
            }
          }
        }
      },
      function (e) {
        console.log('Error: ', e);
      },
      function (){
        console.log('Closed');
      });
  },
  render: function() {
    var state = this.state;
    var regionVersion = this.regionVersion;
    var zoneMap = {};
    _.each(state.nodes, function(n) {
      if (!zoneMap[n.Region])
        zoneMap[n.Region] = {};
      if (!zoneMap[n.Region][n.Zone])
        zoneMap[n.Region][n.Zone] = true;
    })
    var colgroups = _.map(AppConfig.Regions, function(region){
      return <colgroup className="col-result" span={_.keys(zoneMap[region]).length}></colgroup>;
    });
    var regions = _.map(AppConfig.Regions, function(region){
      var len = _.keys(zoneMap[region]).length;
      if (len == 0) len = 1;
      return <th colSpan={len}>{region}</th>;
    });
    var zones = [];
    _.each(AppConfig.Regions, function(region){
      if (_.keys(zoneMap[region]).length == 0) {
        zones.push(<th className="result arch"></th>);
        return;
      }
      _.each(_.keys(zoneMap[region]).sort(), function(zone){
          zones.push(<th className="result arch">{zone}</th>);
      });
    });
    var shardNodes = _.groupBy(state.nodes, function(n) {
      return (n.ParentId == "-") ? n.Id : n.ParentId;
    });
    var shards = _.map(shardNodes, function(nodes){
      var tds = [];
      var master = null;
      // masters
      _.each(nodes, function(node){
        if (node.Role == "master") {
          master = node;
          tds.push(<td>{node.Id.slice(0,10)}</td>);
        }
      });
      // colums
      _.each(AppConfig.Regions, function(region){
        if (_.keys(zoneMap[region]).length == 0) {
          tds.push(<td>-</td>);
          return;
        }
        _.each(_.keys(zoneMap[region]).sort(), function(zone){
          var nn = [];
          _.each(nodes, function(node){
            if (node.Zone == zone) {
              var fail = node.Fail ? "fail":"ok";
              var state = node.State;
              if (node.State == "RUNNING")
                state = "on";
              if (node.State == "OFFLINE")
                state = "off";
              var cls = node.Fail||state!="on" ? "fail": "";
              var role = node.Role == "master" ? "m" : "s";
              var read = node.Readable ? "r":"-";
              var write = node.Writable ? "w":"-";
              if (GC.ShowDetail)
                nn.push(<span className={cls}>({node.Id.slice(0,6)},{read},{write},{role},{fail},{state})</span>);
              else
                nn.push(<span className={cls}>({fail},{state})</span>);
            }
          });
          tds.push(<td>{nn}</td>);
        });
      });
      if (master)
        tds.push(<RangeBar node={master} ranges={master.Ranges} />);
      return <tr>{tds}</tr>;
    });
    return (
      <div>
      <table className="build">
        <colgroup className="col-hash"></colgroup>
        {colgroups}
        <tbody>
          <tr></tr>
          <tr>
            <th></th>
            {regions}
            <th>slots</th>
          </tr>
          <tr>
            <td></td>
            {zones}
            <td></td>
          </tr>
          {shards}
        </tbody>
      </table>
      </div>
    );
  }
});

React.render(
    <BirdView />,
    document.getElementById('content')
);

var CheckBoxes = React.createClass({
  toggleDetail: function(e) {
    GC.ShowDetail = e.target.checked;
  },
  render: function(){
    return (
      <span>
        <input type="checkbox" onChange={this.toggleDetail} />
        show details
      </span>
    );
  }
});

React.render(
    <CheckBoxes />,
    document.getElementById('checkboxes')
);

});
