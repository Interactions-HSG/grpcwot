package main

import (
	"os"

	"github.com/emicklei/proto"
)

var messages map[string]*proto.Message
var rpcFunctions []*linkedRPC

func main() {
	// initialize the messages
	messages = make(map[string]*proto.Message)

	// open sample proto file
	reader, _ := os.Open("files/sample.proto")
	// close reader after the execution finished
	defer func(reader *os.File) {
		err := reader.Close()
		if err != nil {

		}
	}(reader)

	// parse the sample.proto with the protoparser as described on emicklei/proto Github page
	parser := proto.NewParser(reader)
	definition, _ := parser.Parse()

	// walk the proto file and fill messages
	proto.Walk(
		definition,
		proto.WithMessage(addMessage))

	// walk the proto file and fill rpcFunctions
	proto.Walk(definition,
		proto.WithRPC(addRPC))
}

// apply function to be applied on every proto.Message to store it in the messages map
func addMessage(m *proto.Message) {
	messages[m.Name] = m
}

// apply function to be applied on every proto.RPC to generate a linkedRPC and store it in rpcFunctions
func addRPC(rpc *proto.RPC) {
	rpcFunctions = append(rpcFunctions, &linkedRPC{
		rpcFunction: rpc,
		responseMsg: messages[rpc.ReturnsType],
		requestMsg:  messages[rpc.RequestType],
	})
}
