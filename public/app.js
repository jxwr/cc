/// rxjs websocket

var openingObserver = Rx.Observer.create(function() {
  console.log('Opening');
});

var closingObserver = Rx.Observer.create(function() {
  console.log('Socket is about to close');
});

var socket = Rx.DOM.fromWebSocket(
  'ws://127.0.0.1:6201/node/state', 'protocol', openingObserver, closingObserver);

var RxNodeState = socket.map(function(e){ return JSON.parse(e.data); });

/// react

var Node = React.createClass({
  render: function() {
    // console.log("render node", this.props.node);
    var node = this.props.node;
    return (
        <tr className="node">
        <td>{node.Region}</td>
        <td>{node.Fail}</td>
        <td>{node.Readable?"Read":"-"}</td>
        <td>{node.Writable?"Write":"-"}</td>
        <td>{node.Role}</td>
        <td>{node.Ip}:{node.Port}</td>
        <td>{node.Id}</td>
        </tr>
    );
  }
});

var NodeTable = React.createClass({
  componentDidMount: function() {
    var self = this;
    
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
          <Node node={props.nodes[key]} />
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
