package main

import (
	"reflect"
	"strings"

	"github.com/emicklei/proto"
)

// struct to store the linkage between rpcFunction and its *proto.Message s
type linkedRPC struct {
	rpcFunction *proto.RPC
	responseMsg *proto.Message
	requestMsg  *proto.Message
}

// getValueForDefinedPathInRpcFunction returns a value looked up in the linkedRPC if the provided blueprintValue encodes a path to a specific value
// 	if not the original blueprintValue is returned
// @param blueprintValue: is the value that maybe should be searched in the linkedRPC,
// 	for example something like <rpcFunction.Comment.long-title>
// @param rpc: is the linkedRPC as starting point for the search
func getValueForDefinedPathInRpcFunction(blueprintValue string, rpc linkedRPC) string {
	// checks if this is a value that has to be looked up
	if strings.HasPrefix(blueprintValue, "<") &&
		strings.HasSuffix(blueprintValue, ">") {
		var designatedValue string
		// trim the brackets away and divide into single parts like [rpcFunction, Comment, long-title]
		blueprintValue = strings.Trim(blueprintValue, "<>")
		blueprintValueParts := strings.Split(blueprintValue, ".")

		designatedValue = traverseCallParts(
			reflect.Indirect(reflect.ValueOf(rpc)),
			blueprintValueParts,
			0)
		return designatedValue
	} else {
		// returns the original value if it does not have to be looked up
		return blueprintValue
	}
}
