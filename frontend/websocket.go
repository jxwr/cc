package frontend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jxwr/cc/streams"
	"golang.org/x/net/websocket"
)

func nodeStateServer(ws *websocket.Conn) {
	n := 0

	for {
		ns := <-streams.NodeStateStream
		data, err := json.Marshal(ns)
		if err != nil {
			continue
		}
		io.Copy(ws, bytes.NewReader(data))
		n++
		fmt.Println("ws:", n)
	}
}

func RunWebsockServer(bindAddr string) {

	http.Handle("/node/state", websocket.Handler(nodeStateServer))

	err := http.ListenAndServe(bindAddr, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
