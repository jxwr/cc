/// rxjs websocket

var HOST = 'ws://127.0.0.1:6201';
var log = console.log;

var openingObserver = Rx.Observer.create(function() { console.log('Opening socket'); });
var closingObserver = Rx.Observer.create(function() { console.log('Closing socket'); });
var socket = Rx.DOM.fromWebSocket(
  HOST +'/node/state', null, openingObserver, closingObserver);
var RxNodeState = socket.map(function(e){ return JSON.parse(e.data); });

/// react

/// NodeRangeState

var NodeRangeBarItem = React.createClass({
  render: function() {
    var width = 1024;
    var range = this.props.range;
    var style = {
      left: range.Left*width/16384,
      width: (range.Right-range.Left+1)*width/16384,
      backgroundColor: "green"
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
    var items = ranges.map(function (range) {
      return (
          <NodeRangeBarItem range={range} />
      );
    });
    return (
      <tr className="nodeRow">
        <td>{id.substring(0,6)}</td>
        <td>
          <div className="nodeRangeBar">
          {items}
          </div>
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
  render: function() {
    var node = this.props.node;
    var FAIL = node.Fail ? "FAIL":"OK";
    var READ = node.Readable ? "Read":"-";
    var WRITE = node.Writable ? "Write":"-";
    return (
        <tr className="nodeRow">
          <td>{node.State}</td>
          <td>{node.Region}</td>
          <td className={FAIL}>{FAIL}</td>
          <td>{READ}</td>
          <td>{WRITE}</td>
          <td>{node.Role}</td>
          <td>{node.Ip}:{node.Port}</td>
          <td>{node.Id.substring(0,6)}</td>
          <td>{node.Version}</td>
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
  },
  render: function() {
    var nodes = this.props.nodes;
    return (
      <div className="Main">
        <NodeStateTable nodes={nodes} />
        <NodeRangeTable nodes={nodes} />
      </div>
    );
  }
});

React.render(
    <Main nodes={{}} />,
    document.getElementById('content')
);
