package main

import (
	"github.com/emicklei/proto"
)

var messages map[string]*proto.Message
var rpcFunctions []*proto.RPC

func main() {
	// initialize the messages
	messages = make(map[string]*proto.Message)

}
