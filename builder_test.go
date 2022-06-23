package grpcwot

import (
	"errors"
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
	b := &builder{}
	for _, tt := range scalarValueFieldsTests {
		result := b.fieldToDataSchema(tt.in, "")
		if result.DataType != tt.out.DataType {
			t.Errorf("fieldToDataSchema(%v) => \n%v, want \n%v", tt.in, result, tt.out)
		}
	}
}

var messageNameDeterminationTest = []struct {
	in  *proto.Message
	out string
}{
	{
		&proto.Message{Name: "Test"},
		"Test",
	},
	{
		&proto.Message{Name: "Test", Parent: &proto.Message{Name: "ParentTest"}},
		"ParentTest.Test",
	},
	{
		&proto.Message{Name: "Test",
			Parent: &proto.Message{Name: "ParentTest",
				Parent: &proto.Message{Name: "ParentParentTest"}}},
		"ParentParentTest.ParentTest.Test",
	},
	{
		&proto.Message{Name: "Test",
			Parent: &proto.Service{Name: "Service"}},
		"Test",
	},
}

func TestGetFullMessageName(t *testing.T) {
	for _, tt := range messageNameDeterminationTest {
		result := getFullMessageName(tt.in)
		if result != tt.out {
			t.Errorf("getFullMessageName(%v) => \n%v, want \n%v", tt.in, result, tt.out)
		}
	}
}

func errorCheck(t *testing.T, exp error, act error) {
	if act == nil && exp != nil {
		t.Errorf("Expected error %v,\n but no error was raised", exp.Error())
	} else if act != nil && exp == nil {
		t.Errorf("Expected no error, but the follwing error was raised:\n%v", act.Error())
	} else if act != nil && exp != nil && act.Error() != exp.Error() {
		t.Errorf("Expected error message: %v\nbut got: %v", exp.Error(), act.Error())
	}
}

var resolveSingleReferenceTest = []struct {
	in     builder
	inElem refMesTuple
	out    string
	err    error
}{
	{
		// message ParentMessage {
		//   ReferencedMessage testField = 1;
		// }
		// message ReferencedMessage {}
		builder{
			ds: map[string]*wot.DataSchema{
				"ParentMessage":     {},
				"ReferencedMessage": {},
			},
		},
		refMesTuple{pm: "ParentMessage", t: "ReferencedMessage", n: "testField"},
		"ReferencedMessage",
		nil,
	},
	{
		// message Message1 {
		//   message Message2 {
		//     message Message3 {}
		//     Message3 testField = 1;
		//   }
		//   Message2 testField = 1;
		// }
		builder{
			ds: map[string]*wot.DataSchema{
				"Message1":                   {},
				"Message1.Message2":          {},
				"Message1.Message2.Message3": {},
			},
		},
		refMesTuple{pm: "Message1", t: "Message2", n: "testField"},
		"Message1.Message2",
		nil,
	},
	{
		// message Message1 {
		//   message Message2 {
		//     message Message3 {}
		//     Message3 testField = 1;
		//   }
		//   Message2 testField = 1;
		// }
		builder{
			ds: map[string]*wot.DataSchema{
				"Message1":                   {},
				"Message1.Message2":          {},
				"Message1.Message2.Message3": {},
			},
		},
		refMesTuple{pm: "Message1.Message2", t: "Message3", n: "testField"},
		"Message1.Message2.Message3",
		nil,
	},
	{
		// message Message1 {
		//   message Message2 {
		//     message Message3 {
		//       message Message2 {}
		//       Message2.Message3.Message2 testField = 1;
		//       Message5 testField1 = 2;
		//     }
		//     Message3.Message2 testField = 1;
		//   }
		//   Message2.Message3 testField = 1;
		// }
		// message Message5 {}
		builder{
			ds: map[string]*wot.DataSchema{
				"Message1":                            {},
				"Message1.Message2":                   {},
				"Message1.Message2.Message3":          {},
				"Message1.Message2.Message3.Message2": {},
				"Message5":                            {},
			},
		},
		refMesTuple{pm: "Message1.Message2.Message3", t: "Message5", n: "testField"},
		"Message5",
		nil,
	},
	{
		// message Message1 {
		//   message Message2 {
		//     message Message3 {
		//       message Message2 {}
		//       Message2.Message3.Message2 testField = 1;
		//       Message5 testField1 = 2;
		//     }
		//     Message3.Message2 testField = 1;
		//   }
		//   Message2.Message3 testField = 1;
		// }
		// message Message5 {}
		builder{
			ds: map[string]*wot.DataSchema{
				"Message1":                            {},
				"Message1.Message2":                   {},
				"Message1.Message2.Message3":          {},
				"Message1.Message2.Message3.Message2": {},
				"Message5":                            {},
			},
		},
		refMesTuple{pm: "Message1.Message2.Message3", t: "Message2.Message3.Message2", n: "testField"},
		"Message1.Message2.Message3.Message2",
		nil,
	},
	{
		// message Message1 {
		//   message Message2 {
		//     message Message3 {
		//       message Message2 {}
		//       Message2.Message3.Message2 testField = 1;
		//       Message5 testField1 = 2;
		//     }
		//     Message3.Message2 testField = 1;
		//   }
		//   Message2.Message3 testField = 1;
		// }
		// message Message5 {}
		builder{
			ds: map[string]*wot.DataSchema{
				"Message1":                            {},
				"Message1.Message2":                   {},
				"Message1.Message2.Message3":          {},
				"Message1.Message2.Message3.Message2": {},
				"Message5":                            {},
			},
		},
		refMesTuple{pm: "Message1", t: "Message2.Message3", n: "testField"},
		"Message1.Message2.Message3",
		nil,
	},
	{
		// message Message1 {
		//   Message2 testField = 1;
		// }
		builder{
			ds: map[string]*wot.DataSchema{
				"Message1": {},
			},
		},
		refMesTuple{pm: "Message1", t: "Message2", n: "testField"},
		"",
		errors.New("No corresponding message found for type reference Message2 in message Message1"),
	},
	{
		// message Message1 {
		//   message Message2 {
		//     message Message3 {}
		//   }
		//   message Message4 {
		//     Message3 testField = 1;
		//   }
		// }
		builder{
			ds: map[string]*wot.DataSchema{
				"Message1":                   {},
				"Message1.Message2":          {},
				"Message1.Message2.Message3": {},
				"Message1.Message4":          {},
			},
		},
		refMesTuple{pm: "Message1.Message4", t: "Message3", n: "testField"},
		"",
		errors.New("No corresponding message found for type reference Message3 in message Message1.Message4"),
	},
}

func TestResolveSingleReference(t *testing.T) {
	for _, tt := range resolveSingleReferenceTest {
		result, err := tt.in.resolveSingleReference(tt.inElem)

		errorCheck(t, tt.err, err)

		if result != tt.out {
			t.Errorf("Expected %v,\nbut got %v", tt.out, result)
		}
	}
}

type getDataSchemaFunc func(schema map[string]*wot.DataSchema) wot.DataSchema

func getDataSchema(dsInsertMessage string, dsInsertFields []string) getDataSchemaFunc {
	return func(schema map[string]*wot.DataSchema) wot.DataSchema {
		ds := *schema[dsInsertMessage]
		for _, v := range dsInsertFields {
			ds = ds.ObjectSchema.Properties[v]
		}
		return ds
	}
}
func getExpectedSchema(dsExpectedInsertedMessage string) getDataSchemaFunc {
	return getDataSchema(dsExpectedInsertedMessage, []string{})
}

func createInitialMessageDataSchema(props []string) *wot.DataSchema {
	ds := &wot.DataSchema{
		DataType: "object",
		ObjectSchema: &wot.ObjectSchema{
			Properties: map[string]wot.DataSchema{},
		},
	}
	for _, v := range props {
		ds.ObjectSchema.Properties[v] = wot.DataSchema{
			DataType: "object",
		}
	}
	return ds
}
func createInitialMessageDataSchemaOneProperty(prop string) *wot.DataSchema {
	return createInitialMessageDataSchema([]string{prop})
}
func createInitialMessageDataSchemaEmpty() *wot.DataSchema {
	return createInitialMessageDataSchema([]string{})
}

type outRef struct {
	getResultDataSchema   getDataSchemaFunc
	getExpectedDataSchema getDataSchemaFunc
}

func getOutRef(dsInsertMessage string, dsInsertFields []string, dsInsertedMessage string) outRef {
	return outRef{
		getDataSchema(dsInsertMessage, dsInsertFields),
		getExpectedSchema(dsInsertedMessage),
	}
}

var constructMessageNested = []struct {
	in  builder
	out []outRef
	err error
}{
	{
		builder{
			ds: map[string]*wot.DataSchema{
				"Message1": createInitialMessageDataSchemaOneProperty("testField"),
				"Message2": createInitialMessageDataSchemaEmpty(),
			},
			lm: []refMesTuple{
				{pm: "Message1", t: "Message2", n: "testField"},
			},
		},
		[]outRef{
			getOutRef("Message1", []string{"testField"}, "Message2"),
		},
		nil,
	},
	{
		builder{
			ds: map[string]*wot.DataSchema{
				"Message1": createInitialMessageDataSchemaOneProperty("testField"),
				"Message2": createInitialMessageDataSchemaOneProperty("testField"),
				"Message3": createInitialMessageDataSchemaEmpty(),
			},
			lm: []refMesTuple{
				{pm: "Message1", t: "Message2", n: "testField"},
				{pm: "Message2", t: "Message3", n: "testField"},
			},
		},
		[]outRef{
			getOutRef("Message1", []string{"testField"}, "Message2"),
			getOutRef("Message2", []string{"testField"}, "Message3"),
			getOutRef("Message1", []string{"testField", "testField"}, "Message3"),
		},
		nil,
	},
	{
		//	message Message1 {
		//	  message Message2 {
		//	    message Message3 {
		//	      message Message2 {}
		//	      Message3 testField = 1;
		//	      Message5 testField1 = 2;
		//      }
		//	    Message3.Message2 testField = 1;
		//    }
		//	  Message2.Message3 testField = 1;
		//  }
		//	message Message5 {}
		builder{
			ds: map[string]*wot.DataSchema{
				"Message1":                            createInitialMessageDataSchemaOneProperty("testField"),
				"Message1.Message2":                   createInitialMessageDataSchemaOneProperty("testField"),
				"Message1.Message2.Message3":          createInitialMessageDataSchema([]string{"testField", "testField1"}),
				"Message1.Message2.Message3.Message2": createInitialMessageDataSchemaEmpty(),
				"Message5":                            createInitialMessageDataSchemaEmpty(),
			},
			lm: []refMesTuple{
				{pm: "Message1", t: "Message2.Message3", n: "testField"},
				{pm: "Message1.Message2", t: "Message3.Message2", n: "testField"},
				{pm: "Message1.Message2.Message3", t: "Message1.Message2", n: "testField"},
				{pm: "Message1.Message2.Message3", t: "Message5", n: "testField1"},
			},
		},
		[]outRef{
			// Message1 -> Message1.Message2.Message3
			getOutRef("Message1", []string{"testField"}, "Message1.Message2.Message3"),
			// Message1.Message2 -> Message1.Message2.Message3.Message2
			getOutRef("Message1.Message2", []string{"testField"}, "Message1.Message2.Message3.Message2"),
			// Message1.Message2.Message3 -> Message1.Message2
			getOutRef("Message1.Message2.Message3", []string{"testField"}, "Message1.Message2"),
			// Message1.Message2.Message3 -> Message5
			getOutRef("Message1.Message2.Message3", []string{"testField1"}, "Message5"),
			// Message1 -> Message1.Message2.Message3 -> Message5
			getOutRef("Message1", []string{"testField", "testField1"}, "Message5"),
			// Message1 -> Message1.Message2.Message3 -> Message1.Message2
			getOutRef("Message1", []string{"testField", "testField"}, "Message1.Message2"),
			// Message1.Message2.Message3 -> Message1.Message2 -> Message1.Message2.Message3.Message2
			getOutRef("Message1.Message2.Message3", []string{"testField", "testField"},
				"Message1.Message2.Message3.Message2"),
			// Message1 -> Message1.Message2.Message3 -> Message1.Message2 -> Message1.Message2.Message3.Message2
			getOutRef("Message1", []string{"testField", "testField", "testField"},
				"Message1.Message2.Message3.Message2"),
		},
		nil,
	},
	{
		builder{
			ds: map[string]*wot.DataSchema{
				"Message1": createInitialMessageDataSchemaOneProperty("testField"),
				"Message2": createInitialMessageDataSchemaOneProperty("testField"),
			},
			lm: []refMesTuple{
				{pm: "Message1", t: "Message2", n: "testField"},
				{pm: "Message2", t: "Message1", n: "testField"},
			},
		},
		[]outRef{},
		errors.New("proto file contained circle reference in the Messages"),
	},
	{
		builder{
			ds: map[string]*wot.DataSchema{
				"Message1": createInitialMessageDataSchemaOneProperty("testField"),
			},
			lm: []refMesTuple{
				{pm: "Message1", t: "Message2", n: "testField"},
			},
		},
		[]outRef{},
		errors.New("No corresponding message found for type reference Message2 in message Message1"),
	},
	{
		builder{
			ds: map[string]*wot.DataSchema{
				"Message1":                   createInitialMessageDataSchemaOneProperty("testField"),
				"Message1.Message2":          createInitialMessageDataSchemaEmpty(),
				"Message1.Message2.Message3": createInitialMessageDataSchemaEmpty(),
			},
			lm: []refMesTuple{
				{pm: "Message1", t: "Message3", n: "testField"},
			},
		},
		[]outRef{},
		errors.New("No corresponding message found for type reference Message3 in message Message1"),
	},
	{
		builder{
			ds: map[string]*wot.DataSchema{
				"Message1":          createInitialMessageDataSchemaOneProperty("testField"),
				"Message1.Message2": createInitialMessageDataSchemaEmpty(),
				"Message3":          createInitialMessageDataSchemaEmpty(),
				"Message3.Message4": createInitialMessageDataSchemaEmpty(),
			},
			lm: []refMesTuple{
				{pm: "Message1", t: "Message4", n: "testField"},
			},
		},
		[]outRef{},
		errors.New("No corresponding message found for type reference Message4 in message Message1"),
	},
}

func TestConstructMessageNested(t *testing.T) {
	for _, tt := range constructMessageNested {
		err := tt.in.constructMessagesNested()

		errorCheck(t, tt.err, err)

		for _, out := range tt.out {
			result := out.getResultDataSchema(tt.in.ds)
			expected := out.getExpectedDataSchema(tt.in.ds)
			if result.DataType != expected.DataType ||
				result.ObjectSchema != expected.ObjectSchema {
				t.Errorf("constructMessageNested() \n Expected inserted DataSchema \n%v \n but got DataSchema \n%v",
					expected, result)
			}
		}
	}
}
