package main

import (
	"fmt"
	"reflect"
	"strings"
)

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

/*
@param commentTag: the tag to identify which line of a comment should be read
@param commentLines: a reflect.Value which holds a []string with all lines of a comment
This function searches for a line annotated with the comment Tag and returns the value in this line without the annotation
*/
func getCommentLineContent(commentTag string, commentLines reflect.Value) string {
	commentTag = strings.Trim(commentTag, "<>")
	var line string
	for i := 0; i < commentLines.Len(); i++ {
		line = commentLines.Index(i).String()
		// Identifies if the line starts with the searched tag
		if strings.HasPrefix(strings.TrimLeft(line, "\t \n"), "@"+commentTag) {
			// returns the trimmed value of the annotated line without the annotation
			return strings.Trim(
				strings.TrimPrefix(
					strings.TrimLeft(line, "\t \n"),
					"@"+commentTag+":"),
				"\t \n")
		}
	}
	fmt.Println("no comment found for <", commentTag, ">")
	return ""
}
