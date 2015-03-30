$.get('/app/info', function(data){

var AppConfig = data.AppConfig;
var Leader = data.Leader;

var HTTP_HOST = 'http://127.0.0.1:'+Leader.HttpPort;
var WS_HOST = 'ws://127.0.0.1:'+Leader.WsPort;

var openingObserver = Rx.Observer.create(function() { console.log('Opening socket'); });
var closingObserver = Rx.Observer.create(function() { console.log('Closing socket'); });

var stateSocket = Rx.DOM.fromWebSocket(
  WS_HOST +'/node/state', null, openingObserver, closingObserver);
var RxNodeState = stateSocket.map(function(e){ 
  var state = JSON.parse(e.data); 
  delete state.Version;
  delete state.Room;
  delete state.Zone;
  delete state.PFail;
  return state; 
});

var migrateSocket = Rx.DOM.fromWebSocket(
  WS_HOST +'/migrate/state', null, openingObserver, closingObserver);
var RxMigration = migrateSocket.map(function(e){ return JSON.parse(e.data); });

var StateColorMap = {
  "RUNNING": "green",
  "OFFLINE": "yellow",
  "WAIT_FAILOVER_BEGIN": "red",
  "WAIT_FAILOVER_END": "red",
};

var NodeState = React.createClass({
  toggleMode: function(action, perm) {
    $.ajax({
      url: HTTP_HOST+'/node/perm',
      contentType: 'application/json',
      type: "POST",
      data: JSON.stringify({
        node_id: this.props.node.Id,
        action: action,
        perm: perm
      })});
  },
  enableRead: function() {
    this.toggleMode('enable', 'read');
  },
  disableRead: function() {
    this.toggleMode('disable', 'read');
  },
  enableWrite: function() {
    this.toggleMode('enable', 'write');
  },
  disableWrite: function() {
    this.toggleMode('disable', 'write');
  },
  shouldComponentUpdate: function(nextProps, nextState) {
    return !_.isEqual(nextProps.node,this.props.node);
  },
  render: function() {
    var node = this.props.node;
    var role = node.Role == "master" ? "Master" : "Slave";
    var failCol = node.Fail ? "red":"green";
    var stateCol = StateColorMap[node.State];
    var read = node.Readable ? "r":"-";
    var write = node.Writable ? "w":"-";
    var mode = read+"/"+write;
    console.log("render node");
    return (
        <div className="ui card">
        <div className="content">
          <div className="header">
            <i className={failCol+" circle icon"}></i>
            <span className="ui right floated">{role}</span>
            <span>{node.Ip+":"+node.Port}</span>
          </div>
          <div className="meta">
            <span className="ui">{node.Tag}</span>
            <span className="ui right floated">{node.Version}</span>
          </div>
          <div className="description">
            <button onClick={this.enableRead}>+r</button>
            <button onClick={this.disableRead}>-r</button>
            <button onClick={this.enableWrite}>+w</button>
            <button onClick={this.disableWrite}>-w</button>
          </div>
        </div>
        <div>
          <div className={stateCol+" ui bottom left floated label"}>{node.State}</div>
          <div className="ui bottom right floated label">{mode}</div>
        </div>
        </div>
    );
  }
});

var RegionState = React.createClass({
  render: function() {
    var nodes = this.props.nodes;
    var nodeStates = nodes.map(function(n) {
      return <NodeState node={n} />
    });
    return (
      <td>
        {nodeStates}
      </td>
    );
  }
});

var RangeBarItem = React.createClass({
  render: function() {
    var width = parseInt(this.props.width);
    var range = this.props.range;
    var style = {
      left: range.Left*width/16384,
      width: (range.Right-range.Left+1)*width/16384,
      backgroundColor: "#5bbd72",
      height: "10px",
      position: "absolute"
    };
    return (
        <div className="rangeBarItem" style={style}>
        </div>
    );
  }
});

var RangeBar = React.createClass({
  shouldComponentUpdate: function(nextProps, nextState) {
    return !_.isEqual(nextProps.ranges,this.props.ranges);
  },
  render: function() {
    var style = {
      position: "relative",
      width: "640px",
      height: "10px",
      backgroundColor: "lightgrey"
    };
    var id = this.props.nodeId;
    var ranges = this.props.ranges;
    var rangePairs = ranges.map(function (range) {
      return [range.Left,range.Right];
    });
    var rangePairTexts = ranges.map(function (range) {
      return range.Left + "-" + range.Right;
    });
    var items = ranges.map(function (range) {
      return (
          <RangeBarItem range={range} width={style.width} />
      );
    });
    return (
      <div>
        <div>{id}</div>
        <div className="rangeBar" style={style}>
        {items}
        </div>
        <div>{rangePairTexts.join(',')}</div>
      </div>
    );
  }
});

var ReplicaSetState = React.createClass({
  handleRebalance: function(event) {
    event.preventDefault();
    $.ajax({
      url: HTTP_HOST+'/migrate/rebalance',
      contentType: 'application/json',
      type: 'POST',
      data: JSON.stringify({
        method: 'default',
      })});
  },
  render: function() {
    var shard = this.props.shard;
    var coverAllRegions = true;
    var regions = AppConfig.Regions.map(function(region) {
      var nodes = shard.RegionNodes[region];
      if (!nodes) {
        nodes = [];
        coverAllRegions = false;
      }
      return <RegionState nodes={nodes} />;
    });
    var rangeBar = <div>No Master Found</div>;
    // 是否是空节点
    var emptyMaster = null;
    if (shard.Master) {
      master = shard.Master;
      rangeBar = <RangeBar nodeId={master.Id} ranges={master.Ranges} />;
      if ((!master.Fail)&&(master.Ranges.length==0)) {
        emptyMaster = (
          <div>
            <span className="ui small yellow tag label">Empty</span>
            <span className="ui small yellow tag label">
              {coverAllRegions?"CoverAllRegions":"NotCorverAllRegions"}
            </span>
          </div>
        );
      }
    }
    // 能否进行Rebalance（Master,NotDead,CoverAllRegions,NoSlot）
    var rebalanceBtn = null;
    if (emptyMaster && coverAllRegions) {
      rebalanceBtn = <button onClick={this.handleRebalance}>Rebalance</button>;
    }
    return (
        <tr>
        {regions}
        <td>
          {emptyMaster}
          {rangeBar}
          {rebalanceBtn}
        </td>
        </tr>
    );
  }
});

var ClusterState = React.createClass({
  getInitialState: function() {
    return {nodes: {}};
  },
  componentDidMount: function() {
    var self = this;
    RxNodeState.subscribe(
      function (obj) {
        var nodes = self.state.nodes;
        nodes[obj.Id] = obj;
        self.setState({nodes: nodes});
      },
      function (e) {
        console.log('Error: ', e);
      },
      function (){
        console.log('Closed');
      });
  },
  render: function() {
    var headers = AppConfig.Regions.map(function(region) {
      return <th className="four wide">{region}</th>;
    });
    var state = this.state;
    var regionNodes = _.groupBy(state.nodes, function(n) {
      return (n.ParentId == "-") ? n.Id : n.ParentId;
    });
    var shards = _.map(regionNodes, function(nodes) {
      var shard = {Master:null, RegionNodes:{}};
      for (var i = 0; i < nodes.length; i++) {
        var node = nodes[i];
        if (node.Role == "master") {
          shard.Master = node;
        }
        if (!shard.RegionNodes[node.Region]) 
          shard.RegionNodes[node.Region] = [];
        shard.RegionNodes[node.Region].push(node);
      }
      return shard;
    })
    var rows = shards.map(function(shard){
      return <ReplicaSetState shard={shard} />;
    });
    return (
      <div>
      <table className="ui striped green table">
        <thead>
          <tr>
            {headers}
            <th>Slots</th>
          </tr>
        </thead>
        <tbody>
          {rows}
        </tbody>
      </table>
      </div>
    );
  }
});

var AppInfo = React.createClass({
  render: function() {
    var info = this.props.info;
    console.log(info);
    return (
      <table className="ui table">
        <tbody>
        <tr>
          <td>AppName</td>
          <td>MasterRegion</td>
          <td>Regions</td>
          <td>AutoFailover</td>
          <td>Leader Ip</td>
          <td>Leader HttpPort</td>
          <td>Leader WsPort</td>
        </tr>
        <tr>
          <td>{info.AppConfig.AppName}</td>
          <td>{info.AppConfig.MasterRegion}</td>
          <td>{info.AppConfig.Regions.join(",")}</td>
          <td>{info.AppConfig.AutoFailover?"true":"false"}</td>
          <td>{info.Leader.Ip||"127.0.0.1"}</td>
          <td>{info.Leader.HttpPort}</td>
          <td>{info.Leader.WsPort}</td>
        </tr>
        </tbody>
      </table>
    );
  }
});

var MigrationRow = React.createClass({
  render: function() {
    var obj = this.props.obj;
    return (
      <tr>
        <td>{obj.left}-{obj.right}</td>
        <td>{obj.state}</td>
      </tr>
    );
  }
});

var MigrationTable = React.createClass({
  render: function() {
    var task = this.props.task;
    var name = task.SourceId.substring(0,6)+' to '+task.TargetId.substring(0,6);
    var rows = task.Ranges.map(function(range, idx){
      var obj = {left: range.Left, right: range.Right};
      if (idx > task.CurrRangeIndex)
        obj.state = 'Todo';
      if (idx < task.CurrRangeIndex)
        obj.state = 'Done';
      if (idx == task.CurrRangeIndex)
        obj.state = task.State
      return <MigrationRow obj={obj} />;
    });
    return (
      <div className="ui card floated left">
      <table className="ui table">
        <tr>
        <td>{name}</td>
        <td>{task.CurrSlot}</td>
        </tr>
        {rows}
      </table>
      </div>
    );
  }
});

var MigrationPanel = React.createClass({
  getInitialState: function() {
    return {tasks: {}};
  },
  componentDidMount: function() {
    var self = this;
    RxMigration.subscribe(
      function (obj) {
        var tasks = self.state.tasks;
        tasks[obj.SourceId] = obj;
        self.setState({task: tasks});
      },
      function (e) {
        console.log('Error: ', e);
      },
      function (){
        console.log('Closed');
      });
  },  
  render: function() {
    var tasks = this.state.tasks;
    var keys = _.keys(tasks).sort();
    var migs = keys.map(function (key) {
      return (
        <MigrationTable task={tasks[key]} />
      );
    });
    var panel = null;
    if (migs.length > 0)
      panel = <div className="ui segment"> {migs} </div>;
    return panel;
  }
});

var MigrationCtrl = React.createClass({
  handleSubmit: function(event) {
    event.preventDefault();
    var source_id = this.refs['source_id'].getDOMNode().value.trim();
    var target_id = this.refs['target_id'].getDOMNode().value.trim();
    var ranges = this.refs['ranges'].getDOMNode().value.trim();
    ranges = ranges.split(',');
    if (source_id == "" || target_id == "" || ranges.length == 0) return;
    $.ajax({
      url: HTTP_HOST+'/migrate/create',
      contentType: 'application/json',
      type: 'POST',
      data: JSON.stringify({
        source_id: source_id,
        target_id: target_id,
        ranges: ranges
      })});
  },
  render: function() {
    return (
      <div>
        <form className="migrationCtrl" onSubmit={this.handleSubmit}>
        From:<input type="text" ref="source_id" style={{width:"300px"}}/>
        To:<input type="text" ref="target_id" style={{width:"300px"}}/>
        Ranges:<input type="text" ref="ranges"/>
        <button>Migrate</button>
        </form>
      </div>
    );
  }
});

var Main = React.createClass({
  render: function() {
    return (
      <div className="pusher">
        <div className="ui segment">
          <h4>AppInfo</h4>
          <AppInfo info={data} />
        </div>
        <div className="ui segment">
          <h4>Migration</h4>
          <MigrationCtrl />
          <MigrationPanel />
        </div>
        <div className="ui segment">
          <h4>ClusterState</h4>
          <ClusterState />
        </div>
      </div>
    );
  }
});

React.render(
    <Main />,
    document.getElementById('content')
);

});
