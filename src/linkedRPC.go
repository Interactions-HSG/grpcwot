package main

import "github.com/emicklei/proto"

// struct to store the linkage between rpcFunction and its *proto.Message s
type linkedRPC struct {
	rpcFunction *proto.RPC
	responseMsg *proto.Message
	requestMsg  *proto.Message
}
