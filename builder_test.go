package grpcwot

import (
	"testing"

	"github.com/emicklei/proto"
	"github.com/linksmart/thing-directory/wot"
)

// Message parsing test for scalar value fields in fieldToDataSchema()
// Every scalar value type listed for proto3 (https://developers.google.com/protocol-buffers/docs/proto3#scalar) is included with one field
// The test expects, that all scalar value types are converted into the corresponding json value type according to the table provided above
var scalarValueFieldsTests = []struct {
	in  *proto.Field
	out wot.DataSchema
}{
	{
		&proto.Field{Type: "double"},
		wot.DataSchema{DataType: "number"},
	},
	{
		&proto.Field{Type: "float"},
		wot.DataSchema{DataType: "number"},
	},
	{
		&proto.Field{Type: "int32"},
		wot.DataSchema{DataType: "integer"},
	},
	{
		&proto.Field{Type: "int64"},
		wot.DataSchema{DataType: "integer"},
	},
	{
		&proto.Field{Type: "uint32"},
		wot.DataSchema{DataType: "integer"},
	},
	{
		&proto.Field{Type: "uint64"},
		wot.DataSchema{DataType: "integer"},
	},
	{
		&proto.Field{Type: "sint32"},
		wot.DataSchema{DataType: "integer"},
	},
	{
		&proto.Field{Type: "sint64"},
		wot.DataSchema{DataType: "integer"},
	},
	{
		&proto.Field{Type: "fixed32"},
		wot.DataSchema{DataType: "integer"},
	},
	{
		&proto.Field{Type: "fixed64"},
		wot.DataSchema{DataType: "integer"},
	},
	{
		&proto.Field{Type: "sfixed32"},
		wot.DataSchema{DataType: "integer"},
	},
	{
		&proto.Field{Type: "sfixed64"},
		wot.DataSchema{DataType: "integer"},
	},
	{
		&proto.Field{Type: "bool"},
		wot.DataSchema{DataType: "boolean"},
	},
	{
		&proto.Field{Type: "string"},
		wot.DataSchema{DataType: "string"},
	},
	{
		&proto.Field{Type: "bytes"},
		wot.DataSchema{DataType: "string"},
	},
}

func TestScalarFields(t *testing.T) {
	for _, tt := range scalarValueFieldsTests {
		result := fieldToDataSchema(tt.in)
		if result.DataType != tt.out.DataType {
			t.Errorf("fieldToDataSchema(%v) => \n%v, want \n%v", tt.in, result, tt.out)
		}
	}
}
