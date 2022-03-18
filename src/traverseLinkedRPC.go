package main

type AllowedClasses string

// defines classes on which it is allowed to get fields
const (
	LinkedRPC    AllowedClasses = "main.linkedRPC"
	ProtoRPC                    = "proto.RPC"
	ProtoMessage                = "proto.Message"
	ProtoComment                = "proto.Comment"
)

// defines which fields are available for every specific class
var allowedAccess = map[AllowedClasses][]string{
	LinkedRPC: {"rpcFunction", "responseMsg", "requestMsg"},
	ProtoRPC:  {"Position", "Comment", "Name"}, /*"RequestType", "StreamsRequest", "ReturnsType", "StreamsReturns",
	"Elements", "InlineComment", "Parent"*/
	ProtoComment: {"Position", "Lines"}, /*Cstyle, ExtraSlash*/
	ProtoMessage: {"Position", "Comment", "Name" /*, "IsExtend", "Elements", "Parent"*/},
}
