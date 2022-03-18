package main

import "github.com/emicklei/proto"

type linkedRPC struct {
	rpcFunction *proto.RPC
	responseMsg *proto.Message
	requestMsg  *proto.Message
}
