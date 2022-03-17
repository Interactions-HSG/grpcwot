package main

import (
	"github.com/emicklei/proto"
	"os"
)

var messages map[string]*proto.Message
var rpcFunctions []*proto.RPC

func main() {
	// initialize the messages
	messages = make(map[string]*proto.Message)

	// open sample proto file
	reader, _ := os.Open("sample.proto")
	// close reader after the execution finished
	defer func(reader *os.File) {
		err := reader.Close()
		if err != nil {

		}
	}(reader)

	// parse the sample.proto with the protoparser as described on emicklei/proto Github page
	parser := proto.NewParser(reader)
	definition, _ := parser.Parse()

}
