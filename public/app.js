/// rxjs websocket

var HOST = 'ws://127.0.0.1:6201';
var log = console.log;

var openingObserver = Rx.Observer.create(function() { console.log('Opening socket'); });
var closingObserver = Rx.Observer.create(function() { console.log('Closing socket'); });
var socket = Rx.DOM.fromWebSocket(
  HOST +'/node/state', null, openingObserver, closingObserver);
var RxNodeState = socket.map(function(e){ return JSON.parse(e.data); });

/// react

var NodeRow = React.createClass({
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
          <td>{node.Id}</td>
          <td>{node.Version}</td>
        </tr>
    );
  }
});

var NodeTable = React.createClass({
  componentDidMount: function() {
    var self = this;
    
    // 也许该用rx-react之类的Addon
    // 目前逻辑太简单，丑就丑了
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
    var props = this.props;
    var keys = _.keys(props.nodes).sort();
    var nodes = keys.map(function (key) {
      return (
          <NodeRow node={props.nodes[key]} />
      );
    });
    return (
        <table className="nodeTable" nodes="nodes">
          {nodes}
        </table>
    );
  }
});

React.render(
    <NodeTable nodes={{}} />,
    document.getElementById('content')
);
