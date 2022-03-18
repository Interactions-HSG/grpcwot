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

/**
@param object: the object is the actual node where the recursive function is working on in this iteration
@param blueprintValueParts: the slice where the path is stored to access a value
@param actPartsPos: the actual position in this slice
This function runs recursively through the nodes to extract a value for the given blueprintValueParts
*/
func traverseCallParts(object reflect.Value, blueprintValueParts []string, actPartsPos int) string {
	// Checks if the desired value is in the comments and calls getCommentLineContent to extract it
	if AllowedClasses(object.Type().String()) == ProtoComment &&
		!contains(allowedAccess[ProtoComment], blueprintValueParts[actPartsPos]) {
		return getCommentLineContent(
			blueprintValueParts[actPartsPos],
			object.FieldByName("Lines"),
		)
	}
	if _, ok := allowedAccess[AllowedClasses(object.Type().String())]; !ok {
		// Checks if the objects type is allowed
		fmt.Println("The following struct is not listed in the allowed access:", object.Type().String())
		return ""
	} else if !contains(allowedAccess[AllowedClasses(object.Type().String())], blueprintValueParts[actPartsPos]) {
		// Checks if the access of the actual blueprint values is possible on this object
		fmt.Println("The function", blueprintValueParts[actPartsPos],
			"is not listed in allowed classes for", AllowedClasses(object.Type().String()))
		return ""
	}
	// Extracts the desired field out of the object and reassigns it to the object
	object = object.FieldByName(blueprintValueParts[actPartsPos])

	// Checks if the blueprintValueParts are not completely processed
	if actPartsPos < len(blueprintValueParts)-1 {
		// reflect.Pointer is dereferenced to allow field access in next recursive call
		if object.Kind() == reflect.Pointer {
			object = object.Elem()
		}
		// recursive call on the newly retrieved object and the next part of the blueprintValue
		return traverseCallParts(
			object,
			blueprintValueParts,
			actPartsPos+1,
		)
	}

	// when the blueprintValueParts are completely processed the fields kind is examined format the value correctly
	switch object.Kind() {
	case reflect.Bool:
		return fmt.Sprintf("%t", object.Bool())
	case reflect.Int:
		return fmt.Sprintf("%d", object.Int())
	case reflect.String:
		return fmt.Sprintf("%s", object.String())
	case reflect.Float64, reflect.Float32:
		return fmt.Sprintf("%f", object.Float())
	default:
		fmt.Println("The final blueprintValue is not bool, int, string or float type and cannot be printed")
		return ""
	}
}

/*
Small helper function to determine if a slice of strings contains a specific string
*/
func contains(arr []string, s string) bool {
	for _, v := range arr {
		if s == v {
			return true
		}
	}
	return false
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
