$.get('/app/info', function(data){

var AppConfig = data.AppConfig;
var Leader = data.Leader;

var HTTP_HOST = 'http://'+Leader.Ip+':'+Leader.HttpPort;
var WS_HOST = 'ws://'+Leader.Ip+':'+Leader.WsPort;

var GlobalNodes = [];
var GC = {
  ShowMigrateActions: false,
  ShowAdvanceNodeActions: false,
};

if (location.origin != HTTP_HOST)
  location = HTTP_HOST + location.pathname;

var openingObserver = Rx.Observer.create(function() { console.log('Opening socket'); });
var closingObserver = Rx.Observer.create(function() { console.log('Closing socket'); });

var migrateSocket = Rx.DOM.fromWebSocket(
  WS_HOST +'/migrate/state', null, openingObserver, closingObserver);
var RxMigration = migrateSocket.map(function(e){ return JSON.parse(e.data); });

var stateSocket = Rx.DOM.fromWebSocket(
  WS_HOST +'/node/state', null, openingObserver, closingObserver);
var RxNodeState = stateSocket.map(function(e){ 
  var state = JSON.parse(e.data); 
  delete state.Room;
  delete state.Zone;
  delete state.PFail;
  delete state.ClusterStatsMessagesSent;
  delete state.ClusterStatsMessagesReceived;
  delete state.Expires;
  return state;
});

var StateColorMap = {
  "RUNNING": "green",
  "OFFLINE": "yellow",
  "WAIT_FAILOVER_BEGIN": "red",
  "WAIT_FAILOVER_END": "red",
};

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
    var name = task.SourceId.substring(0,10)+' to '+task.TargetId.substring(0,10);
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
      <table className="migrate" style={{fontSize:"20px"}}>
        <tr>
        <th>{name}</th>
        <th>{task.CurrSlot}</th>
        </tr>
        {rows}
      </table>
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
    if (migs.length == 0)
      migs = "No MigTask";
    var style = {
      position: "fixed",
      right: "20px",
      top: "20px",
      zIndex: 10
    };
    var panel = (
        <div className="taskPanel" style={style}>
        <table>
        <tr>
          <td className="mig-task-panel">{migs}</td>
          <td><button onClick={this.toggle}>#</button></td>
        </tr>
        </table>
        </div>
    );
    return panel;
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
      return <option key={n.Id} value={n.Id}>{n.Id.slice(0,10)} - {n.Ip}:{n.Port}</option>;
    }).filter(function(n) { return n != null; });
    var style = {padding:"0 20px",color:"black",display:GC.ShowMigrateActions?"":"none"};
    return (
       <span style={style}>
         <span>
           <label>migrate ranges</label>
           <input ref={source.Id+"_ranges"} type="text"/>
         </span>
         <span>
           <label>to</label>
           <select ref={source.Id+"_target"}>{idOptions}</select>
         </span>
         <button onClick={this.handleCreateMigrateTask}>Migrate</button>
       </span>
    );
  }
});

var NodeAction = React.createClass({
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
  handleMeet: function() {
    $.ajax({
      url: HTTP_HOST+'/node/meet',
      contentType: 'application/json',
      type: "POST",
      data: JSON.stringify({
        node_id: this.props.node.Id,
      })});
  },
  handleForget: function() {
    $.ajax({
      url: HTTP_HOST+'/node/forgetAndReset',
      contentType: 'application/json',
      type: "POST",
      data: JSON.stringify({
        node_id: this.props.node.Id,
      })});
  },
  handleReparent: function() {
    var node = this.props.node;
    var select = this.refs[node.Id+'_reparent'].getDOMNode();
    var targetId = select.options[select.selectedIndex].value;
    $.ajax({
      url: HTTP_HOST+'/node/replicate',
      contentType: 'application/json',
      type: "POST",
      data: JSON.stringify({
        child_id: node.Id,
        parent_id: targetId
      })});
  },
  handleSetAsMaster: function() {
    $.ajax({
      url: HTTP_HOST+'/node/setAsMaster',
      contentType: 'application/json',
      type: "POST",
      data: JSON.stringify({
        node_id: this.props.node.Id,
      })});
  },
  handleFailoverTakeover: function() {
    $.ajax({
      url: HTTP_HOST+'/failover/takeover',
      contentType: 'application/json',
      type: "POST",
      data: JSON.stringify({
        node_id: this.props.node.Id,
      })});
  },
  render: function() {
    var node = this.props.node;
    var options = _.map(GlobalNodes, function(n) {
      if (n.Id == node.Id) return null;
      return <option key={n.Id} value={n.Id}>{n.Ip}:{n.Port}</option>;
    }).filter(function(n) { return n != null; });
    var reparent = (
      <span>
        <button onClick={this.handleReparent}>RP</button>
        <select ref={node.Id+"_reparent"}>{options}</select>
      </span>
    );
    return (
      <td className="action">
        <button onClick={this.enableRead}>+r</button>
        <button onClick={this.disableRead}>-r</button>
        <button onClick={this.enableWrite}>+w</button>
        <button onClick={this.disableWrite}>-w</button>
        <span style={{display:GC.ShowAdvanceNodeActions?"":"none"}}>
        <button onClick={this.handleMeet}>Meet</button>
        <button onClick={this.handleForget}>Forget&Reset</button>
        <button onClick={this.handleSetAsMaster}>SetAsMaster</button>
        <button onClick={this.handleFailoverTakeover}>Takeover</button>
        {reparent}
        </span>
      </td>
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
  header: (
      <tr>
        <th></th>
        <th>ver</th>
        <th>id</th>
        <th>tag</th>
        <th>ip:port</th>
        <th>role</th>
        <th>mode</th>
        <th>fail</th>
        <th>state</th>
        <th>dump</th>
        <th>repl</th>
        <th>keys</th>
        <th>qps</th>
        <th>net_in</th>
        <th>net_out</th>
        <th>mem_used</th>
        <th>link</th>
        <th>slots</th>
      </tr>
  ),
  render: function() {
    var shard = this.props.shard;
    var masterNodes = this.props.masterNodes;
    var coverAllRegions = true;
    var allNodes = [];
    // first row is master
    if (shard.Master) allNodes.push(shard.Master);
    _.each(AppConfig.Regions, function(region) {
      var nodes = shard.RegionNodes[region];
      if (!nodes) {
        nodes = [];
        coverAllRegions = false;
      }
      _.each(nodes, function(n){
        if (n != null && n != shard.Master)
          allNodes.push(n);
      });
    });
    var rows = _.map(allNodes, function(node){
      var read = node.Readable ? "r":"-";
      var write = node.Writable ? "w":"-";
      var mode = read+"/"+write;
      var fail = node.Fail ? "fail":"ok";
      var ranges = node.Ranges;
      var rangePairs = ranges.map(function (range) {
        return [range.Left,range.Right];
      });
      var rangePairTexts = ranges.map(function (range) {
        return range.Left + "-" + range.Right;
      });
      return (
          <tr>
            <NodeAction node={node} />
            <td>{node.Version}</td>
            <td>{node.Id.slice(0,10)}</td>
            <td>{node.Tag}</td>
            <td>{node.Ip}:{node.Port}</td>
            <td>{node.Role}</td>
            <td>{mode}</td>
            <td className={fail}>{fail}</td>
            <td className={node.State!="RUNNING"?"fail":""}>{node.State}</td>
            <td>{node.RdbBgsaveInProgress?"bgsaving":"-"}</td>
            <td>{node.ReplOffset}</td>
            <td>{node.Keys}</td>
            <td>{node.InstantaneousOpsPerSec}</td>
            <td>{node.InstantaneousInputKbps.toFixed(2)}Kbps</td>
            <td>{node.InstantaneousOutputKbps.toFixed(2)}Kbps</td>
            <td>{(node.UsedMemory/1024.0/1024.0/1024.0).toFixed(3)}G</td>
            <td>{node.Role=="slave"?node.MasterLinkStatus:"-"}</td>
            <td>{rangePairTexts.join(',')}</td>
          </tr>
      );
    });
    return (
      <div>
        <h4>
        <small>Replicas ({shard.Master ? shard.Master.Id.slice(0,10) : '-'})</small>
        <MigrationForm sourceMaster={shard.Master} masterNodes={masterNodes} />
        </h4>
        <table>
          <tbody>
          <tr></tr>
          {this.header}
          {rows}
          </tbody>
        </table>
      </div>
    );
  }
});

function IsStandbyNode(shard, node) {
  if (node.Role != "master") return false;
  if (node.Free) return false;
  if (node.Ranges.length > 0) return false;
  if (_.isEqual(_.keys(shard.RegionNodes).sort(),AppConfig.Regions.sort())) return false;
  if (_.flatten(_.values(shard.RegionNodes)) > 1) return false;
  return true;
}

var ClusterState = React.createClass({
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
    var regionNodes = _.groupBy(state.nodes, function(n) {
      return (n.ParentId == "-") ? n.Id : n.ParentId;
    });
    var masterNodes = [];
    var freeNodes = [];
    var shards = _.map(regionNodes, function(nodes) {
      var shard = {Master:null, RegionNodes:{}};
      for (var i = 0; i < nodes.length; i++) {
        var node = nodes[i];
        if (node.Free) 
          freeNodes.push(node);
        if (node.Role == "master") {
          shard.Master = node;
        }
        if (!shard.RegionNodes[node.Region]) 
          shard.RegionNodes[node.Region] = [];
        shard.RegionNodes[node.Region].push(node);
        if (node.Version) regionVersion[node.Region] = node.Version;
      }
      return shard;
    })
    var standbyNodes = [];
    var standbyNodeTable = null;
    var onlineMasters = [];
    var onlineShards = _.filter(shards, function(shard) {
      var master = shard.Master;
      if (!master) return true;
      if (master.Free) return false;
      if (IsStandbyNode(shard, master)) {
        standbyNodes.push(master);
        master.Standby = true;
        return false;
      } else {
        onlineMasters.push(master);
        master.Standby = false;
        return true;
      }
    })
    var rss = onlineShards.map(function(shard) {
      return <ReplicaSetState shard={shard} masterNodes={onlineMasters} />;
    });
    return (
      <div>
        {rss}
        <MigrationTaskPanel />
      </div>
    );
  }
});

React.render(
    <ClusterState />,
    document.getElementById('content')
);

var CheckBoxes = React.createClass({
  toggleNodeActions: function(e) {
    GC.ShowAdvanceNodeActions = e.target.checked;
  },
  toggleMigActions: function(e) {
    GC.ShowMigrateActions = e.target.checked;
  },
  render: function(){
    return (
      <span>
        <input type="checkbox" onChange={this.toggleNodeActions}/>
        node actions
        <input type="checkbox" onChange={this.toggleMigActions}/>
        migrate actions
      </span>
    );
  }
});

React.render(
    <CheckBoxes />,
    document.getElementById('checkboxes')
);


});
