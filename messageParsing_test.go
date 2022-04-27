package grpcwot

import (
	"encoding/json"
	"github.com/linksmart/thing-directory/wot"
	"testing"
)

func TestMessageParserStandardMessage(t *testing.T) {
	MessageParserTestHelper(t, "standardMessage")
}

func TestMessageParserLinkedMessages(t *testing.T) {
	MessageParserTestHelper(t, "linkedMessage")
}

func TestMessageParserMultipleMessages(t *testing.T) {
	MessageParserTestHelper(t, "multipleMessages")
}

func TestMessageParserNestedMessage(t *testing.T) {
	MessageParserTestHelper(t, "nestedMessage")
}

func TestMessageParserComplexNestedMessage(t *testing.T) {
	MessageParserTestHelper(t, "complexNestedMessage")
}

func MessageParserTestHelper(t *testing.T, f string) {
	a := parseProtoFileToBuilder("testFiles/testMessages/protos/" + f + ".proto").ds
	e := unmarshalExpectedFileForDataSchemas("testFiles/testMessages/jsonLD/" + f + ".json")
	//x, _ := json.Marshal(a)
	//fmt.Println(x)
	checkDataSchemaMapEquality(t, e, a)
}

func checkDataSchemaMapEquality(t *testing.T, e, a map[string]*wot.DataSchema) {
	if len(e) != len(a) {
		t.Errorf("The number of messages differ for expected and actual. "+
			"Number of expected is %d, number of actual is %d", len(e), len(a))
	}
	for k, v := range e {
		v2, ok := a[k]
		if !ok {
			t.Errorf("The actual message map does not contain the expected key %s", k)
		} else {
			checkDataSchemaEquality(t, *v, *v2)
		}
	}
}

func checkDataSchemaPropertiesMapEquality(t *testing.T, e, a map[string]wot.DataSchema) {
	if len(e) != len(a) {
		t.Errorf("The length of expected and actual property map differ. "+
			"Length of expected is %d, length of actual is %d", len(e), len(a))
	}
	for k, v := range e {
		v2, ok := a[k]
		if !ok {
			t.Errorf("The actual property map does not contain the expected key %s", k)
		} else {
			checkDataSchemaEquality(t, v, v2)
		}
	}
}

func checkDataSchemaEquality(t *testing.T, e, a wot.DataSchema) {
	if e.DataType != a.DataType {
		t.Errorf("Datatypes of expected and actual do not match. "+
			"Expected DataType was %s and actual DataType is %s", e.DataType, a.DataType)
	}
	switch e.DataType {
	case "object":
		checkDataSchemaPropertiesMapEquality(t, e.ObjectSchema.Properties, a.ObjectSchema.Properties)
	}
}

func getBuilderDataSchemasAsJSON(b *builder) (j []byte) {
	v, err := json.Marshal(b.ds)
	if err != nil {
		return []byte{}
	}
	return v
}
