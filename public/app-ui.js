$.get('/app/info', function(data){

var AppConfig = data.AppConfig;
var Leader = data.Leader;

var HTTP_HOST = 'http://'+Leader.Ip+':'+Leader.HttpPort;
var WS_HOST = 'ws://'+Leader.Ip+':'+Leader.WsPort;

console.log("HttpHost", HTTP_HOST);

var openingObserver = Rx.Observer.create(function() { console.log('Opening socket'); });
var closingObserver = Rx.Observer.create(function() { console.log('Closing socket'); });

var stateSocket = Rx.DOM.fromWebSocket(
  WS_HOST +'/node/state', null, openingObserver, closingObserver);
var RxNodeState = stateSocket.map(function(e){ 
  var state = JSON.parse(e.data); 
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
  render: function() {
    var node = this.props.node;
    var role = node.Role == "master" ? "Master" : "Slave";
    var failCol = node.Fail ? "red":"green";
    var stateCol = StateColorMap[node.State];
    var read = node.Readable ? "r":"-";
    var write = node.Writable ? "w":"-";
    var mode = read+"/"+write;
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
      return <NodeState key={n.Id} node={n} />
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
  render: function() {
    var style = {
      position: "relative",
      width: "640px",
      height: "10px",
      backgroundColor: "lightgrey"
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
      <div className="ui segment">
        <div className="ui top left attached label">{node.Id} - {node.Ip}:{node.Port}</div>
        <div className="rangeBar" style={style}>
        {items}
        </div>
        <div>{rangePairTexts.join(',')}</div>
      </div>
    );
  }
});

var MigrationForm = React.createClass({
  toggle: function() {
    var m = this.props.sourceMaster;
    if(!m) return;
    $('.'+m.Id+'.mig.segment').toggle();
  },
  handleCreateMigrateTask: function() {
    var m = this.props.sourceMaster;
    if(!m) return;
    var select = this.refs[m.Id+'_target'].getDOMNode();
    var targetId = select.options[select.selectedIndex].value;
    var ranges = this.refs[m.Id+'_ranges'].getDOMNode().value.trim();
    ranges = ranges.split(',');
    if (targetId == "" || ranges.length == 0) return;
    console.log(targetId, ranges);
    $.ajax({
      url: HTTP_HOST+'/migrate/create',
      contentType: 'application/json',
      type: 'POST',
      data: JSON.stringify({
        source_id: m.Id,
        target_id: targetId,
        ranges: ranges
      })});
  },
  render: function() {
    var source = this.props.sourceMaster;
    if (!source) return <div></div>;
    var masterNodes = this.props.masterNodes;
    var idOptions = masterNodes.map(function(n) {
      if (!source || n.Id == source.Id) return null;
      return <option key={n.Id} value={n.Id}>{n.Id} - {n.Ip}:{n.Port}</option>;
    }).filter(function(n) { return n != null; });
    return (
        <div>
          <div className="compact ui icon button" onClick={this.toggle}>
            <i className="content icon"></i>
          </div>
          <div className={source.Id+" mig ui segment"} style={{display:"none"}}>
            <div className="ui form">
              <div className="field">
                <label>Migrate ranges</label>
                <input ref={source.Id+"_ranges"} type="text"/>
              </div>
              <div className="field">
                <label>to</label>
                <select ref={source.Id+"_target"}>{idOptions}</select>
              </div>
              <div className="ui tiny blue submit button"
                   onClick={this.handleCreateMigrateTask}>Migrate</div>
            </div>
          </div>
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
  // For optimizing
  shouldComponentUpdate: function(nextProps, nextState) {
    return !(_.isEqual(nextProps.shard, this.props.shard) && 
             _.isEqual(nextProps.masterNodes, this.props.masterNodes));
  },
  render: function() {
    var shard = this.props.shard;
    var masterNodes = this.props.masterNodes;
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
    var tags = null;
    var emptyMaster = false;
    if (shard.Master) {
      master = shard.Master;
      rangeBar = <RangeBar node={master} ranges={master.Ranges} />;
      if ((!master.Fail)&&(master.Ranges.length==0)) {
        emptyMaster = true;
        tags = (
          <span>
            <span className="compat ui yellow tag label">Empty</span>
            <span className="compat ui yellow tag label">
              {coverAllRegions?"CoverAllRegions":"NotCorverAllRegions"}
            </span>
          </span>
        );
      }
    }
    // 能否进行Rebalance（Master,NotDead,CoverAllRegions,NoSlot）
    var rebalanceBtn = null;
    if (emptyMaster && coverAllRegions) {
      rebalanceBtn = <button className="blue compact tiny ui button" 
                             onClick={this.handleRebalance}>Rebalance To This Node</button>;
    }
    return (
        <tr>
        {regions}
        <td>
          {tags} {rebalanceBtn}
          {rangeBar}
          <MigrationForm sourceMaster={shard.Master} masterNodes={masterNodes} />
        </td>
        </tr>
    );
  }
});

var ClusterState = React.createClass({
  regionVersion: {},
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
    var state = this.state;
    var regionVersion = this.regionVersion;
    var regionNodes = _.groupBy(state.nodes, function(n) {
      return (n.ParentId == "-") ? n.Id : n.ParentId;
    });
    var masterNodes = [];
    var shards = _.map(regionNodes, function(nodes) {
      var shard = {Master:null, RegionNodes:{}};
      for (var i = 0; i < nodes.length; i++) {
        var node = nodes[i];
        if (node.Role == "master") {
          shard.Master = node;
          masterNodes.push(node);
        }
        if (!shard.RegionNodes[node.Region]) 
          shard.RegionNodes[node.Region] = [];
        shard.RegionNodes[node.Region].push(node);
        if (node.Version) regionVersion[node.Region] = node.Version;
        // For optimizing
        delete node.Version;
      }
      return shard;
    })
    var rows = shards.map(function(shard){
      return <ReplicaSetState shard={shard} masterNodes={masterNodes} />;
    });
    var headers = AppConfig.Regions.map(function(region) {
      return <th className="four wide">{region}({regionVersion[region]})</th>;
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

var MigrationTask = React.createClass({
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

var MigrationTaskTable = React.createClass({
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
      return <MigrationTask obj={obj} />;
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

var MigrationTaskPanel = React.createClass({
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
  toggle: function() {
    $(".mig-task-panel").toggle();
  },
  render: function() {
    var tasks = this.state.tasks;
    var keys = _.keys(tasks).sort();
    var migs = keys.map(function (key) {
      return (
        <MigrationTaskTable task={tasks[key]} />
      );
    });
    var style = {
      position: "fixed",
      left: "20px",
      top: "20px",
      zIndex: 10
    };
    var panel = null;
    if (migs.length > 0)
      panel = (
        <div className="taskPanel ui purple inverted segment" style={style}>
          <a className="ui left corner label" onClick={this.toggle}>
            <i className="content icon"></i>
          </a>
          <div className="mig-task-panel">{migs}</div>
        </div>
      );
    return panel;
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
          <h4>ClusterState</h4>
          <ClusterState />
        </div>
        <MigrationTaskPanel />
      </div>
    );
  }
});

React.render(
    <Main />,
    document.getElementById('content')
);

});
