/// rxjs websocket

var HTTP_HOST = 'http://127.0.0.1:6200';
var WS_HOST = 'ws://127.0.0.1:6201';
var log = console.log;

var openingObserver = Rx.Observer.create(function() { console.log('Opening socket'); });
var closingObserver = Rx.Observer.create(function() { console.log('Closing socket'); });

var stateSocket = Rx.DOM.fromWebSocket(
  WS_HOST +'/node/state', null, openingObserver, closingObserver);
var RxNodeState = stateSocket.map(function(e){ return JSON.parse(e.data); });

var migrateSocket = Rx.DOM.fromWebSocket(
  WS_HOST +'/migrate/state', null, openingObserver, closingObserver);
var RxMigration = migrateSocket.map(function(e){ return JSON.parse(e.data); });

/// react

var MigratingRow = React.createClass({
  render: function() {
    var obj = this.props.obj;
    return (
      <tr className="nodeRow">
        <td>{obj.left}-{obj.right}</td>
        <td>{obj.state}</td>
      </tr>
    );
  }
});

var MigrationTable = React.createClass({
  render: function() {
    var mig = this.props.mig;
    var name = mig.SourceId.substring(0,6)+' to '+mig.TargetId.substring(0,6);
    var rows = mig.Ranges.map(function(range, idx){
      var obj = {left: range.Left, right: range.Right};
      if (idx > mig.CurrRangeIndex)
        obj.state = "Todo";
      if (idx < mig.CurrRangeIndex)
        obj.state = "Done";
      if (idx == mig.CurrRangeIndex)
        obj.state = mig.State
      return <MigratingRow obj={obj} />;
    });
    return (
      <div className="migrationTable">
        <table className="nodeTable">
        <tr className="nodeRow">
        <td>{name}</td>
        <td>{mig.CurrSlot}</td>
        </tr>
        {rows}
      </table>
      </div>
    );
  }
});

var MigrationPanel = React.createClass({
  render: function() {
    var migMap = this.props.migMap;
    var keys = _.keys(migMap).sort();
    var migs = keys.map(function (key) {
      return (
        <MigrationTable mig={migMap[key]} />
      );
    });
    return (
      <div className="migrationPanel">
        {migs}
      </div>
    );
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
      type: "POST",
      data: JSON.stringify({
        source_id: source_id,
        target_id: target_id,
        ranges: ranges
      })});
  },
  render: function() {
    return (
      <form className="migrationCtrl" onSubmit={this.handleSubmit}>
        From:<input type="text" ref="source_id"/>
        To:<input type="text" ref="target_id"/>
        Ranges:<input type="text" ref="ranges"/>
        <button>Migrate</button>
      </form>
    );
  }
});

/// NodeRangeState

var NodeRangeBarItem = React.createClass({
  render: function() {
    var width = 1024;
    var range = this.props.range;
    var style = {
      left: range.Left*width/16384,
      width: (range.Right-range.Left+1)*width/16384,
      backgroundColor: "#00BB00"
    };
    return (
        <div className="nodeRangeBarItem" style={style}>
        </div>
    );
  }
});

var NodeRangeRow = React.createClass({
  render: function() {
    var id = this.props.nodeid;
    var ranges = this.props.ranges;
    var rangePairs = ranges.map(function (range) {
      return [range.Left,range.Right];
    });
    var items = ranges.map(function (range) {
      return (
          <NodeRangeBarItem range={range} />
      );
    });
    return (
      <tr className="nodeRow">
        <td>{id.substring(0,6)}</td>
        <td>
          <div>{id}</div>
          <div className="nodeRangeBar">
          {items}
          </div>
          <div>{JSON.stringify(rangePairs)}</div>
        </td>
      </tr>
    );
  }
});

var NodeRangeTable = React.createClass({
  render: function() {
    var nodes = this.props.nodes;
    var keys = _.keys(this.props.nodes).filter(function(key){
      return nodes[key].Role == "master";
    }).sort();
    var rows = keys.map(function (key) {
      return (
        <NodeRangeRow nodeid={nodes[key].Id} ranges={nodes[key].Ranges} />
      );
    });
    return (
      <table className="nodeTable">
        {rows}
      </table>
    )
  }
});

/// NodeStateTable

var NodeStateRow = React.createClass({
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
    var FAIL = node.Fail ? "FAIL":"OK";
    var READ = node.Readable ? "Read":"-";
    var WRITE = node.Writable ? "Write":"-";
    var EMPTY = (node.Role=="master")&&(!node.Fail)&&(node.Ranges.length==0) ? "EmptyMaster":"-";
    return (
        <tr className="nodeRow">
          <td>{node.State}</td>
          <td>{node.Region}</td>
          <td className={FAIL}>{FAIL}</td>
          <td>{READ}</td>
          <td>{WRITE}</td>
          <td>{node.Role}</td>
          <td>{node.Ip}:{node.Port}</td>
          <td>{node.Id}</td>
          <td>{EMPTY}</td>
          <td>{node.Version}</td>
          <td>
            <button onClick={this.enableRead}>+r</button>
            <button onClick={this.disableRead}>-r</button>
            <button onClick={this.enableWrite}>+w</button>
            <button onClick={this.disableWrite}>-w</button>
          </td>
        </tr>
    );
  }
});

var NodeStateTable = React.createClass({
  render: function() {
    var props = this.props;
    var keys = _.keys(props.nodes).sort();
    var nodes = keys.map(function (key) {
      return (
          <NodeStateRow node={props.nodes[key]} />
      );
    });
    return (
        <table className="nodeTable">
          {nodes}
        </table>
    );
  }
});

var Main = React.createClass({
  componentDidMount: function() {
    var self = this;
    // 也许该用rx-react之类的Addon
    RxNodeState.subscribe(
      function (obj) {
        var nodes = self.props.nodes;
        nodes[obj.Id] = obj;
        self.setState({nodes: nodes});
      },
      function (e) {
        console.log('Error: ', e);
      },
      function (){
        console.log('Closed');
      });

    RxMigration.subscribe(
      function (obj) {
        var migMap = self.props.migMap;
        migMap[obj.SourceId] = obj;
        self.setState({migMap: migMap});
      },
      function (e) {
        console.log('Error: ', e);
      },
      function (){
        console.log('Closed');
      });
  },
  render: function() {
    var nodes = this.props.nodes;
    var migMap = this.props.migMap;
    return (
      <div className="Main">
        <NodeStateTable nodes={nodes} />
        <MigrationCtrl />
        <MigrationPanel migMap={migMap} />
        <NodeRangeTable nodes={nodes} />
      </div>
    );
  }
});

React.render(
    <Main nodes={{}} migMap={{}} />,
    document.getElementById('content')
);
