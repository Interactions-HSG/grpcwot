package main

import (
	"github.com/emicklei/proto"
	"reflect"
	"strings"
)

// struct to store the linkage between rpcFunction and its *proto.Message s
type linkedRPC struct {
	rpcFunction *proto.RPC
	responseMsg *proto.Message
	requestMsg  *proto.Message
}

/**
@param blueprintValue: is the value that maybe should be searched in the linkedRPC,
	for example something like <rpcFunction.Comment.long-title>
@param rpc: is the linkedRPC as starting point for the search
This function returns a value looked up in the linkedRPC based on the given path if one was provided
	or returns the original value if not
*/
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
