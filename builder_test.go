package grpcwot

import (
	"encoding/json"
	"fmt"
	"github.com/emicklei/proto"
	"github.com/linksmart/thing-directory/wot"
	"io/ioutil"
	"os"
	"testing"
)

type unmarshalExpected func([]byte) error

// Message parsing tests - scalar field types
// Proto files as input files are provided under testFiles/testScalarFieldTypes/proto
// The desired structure of data schemas are provided under testFiles/testScalarFieldTypes/jsonLD

// The input file scalarFieldTypeMessage provides only one message with fields of scalar value types
// Every scalar value type listed for proto3 (https://developers.google.com/protocol-buffers/docs/proto3#scalar) is included with one field
// The test expects, that all scalar value types are converted into the corresponding json value type according to the table provided above
func TestMessageParserStandardMessage(t *testing.T) {
	MessageParserTestHelper(t, "scalarFieldTypeMessage", "testFiles/testScalarFieldTypes")
}

// Helper functions for message testing

func MessageParserTestHelper(t *testing.T, f string, dir string) {
	a := parseProtoFileToBuilder(dir + "/protos/" + f + ".proto").ds
	e := map[string]*wot.DataSchema{}
	unmarshalExpectedFile(dir+"/json/"+f+".json", func(b []byte) error {
		return json.Unmarshal(b, &e)
	})
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

func checkDataSchemaEquality(t *testing.T, e, a wot.DataSchema) {
	if e.DataType != a.DataType {
		t.Errorf("Datatypes of expected and actual do not match. "+
			"Expected DataType was %s and actual DataType is %s", e.DataType, a.DataType)
	}
	if e.DataType == "object" {
		checkDataSchemaPropertiesMapEquality(t, e.ObjectSchema.Properties, a.ObjectSchema.Properties)
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

// General helper functions

func parseProtoFileToBuilder(protoFile string) (b *builder) {
	b = newBuilder("", 0)

	// parse the protoFile with the emicklei/proto
	reader, _ := os.Open(protoFile)
	defer reader.Close()
	parser := proto.NewParser(reader)
	definition, err := parser.Parse()
	if err != nil {
		return nil
	}

	proto.Walk(definition,
		proto.WithMessage(b.HandleMessage))

	return b
}

func unmarshalExpectedFile(expectedFile string, fn unmarshalExpected) {
	jsonFile, err := os.Open(expectedFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func(jsonFile *os.File) {
		jsonFile.Close()
	}(jsonFile)

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return
	}
	err = fn(byteValue)
	if err != nil {
		return
	}
}

// readByteValueFromJsonFile reads in a json file into byteValue
func readByteValueFromJsonFile(file string) ([]byte, error) {
	jsonFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer func(jsonFile *os.File) {
		jsonFile.Close()
	}(jsonFile)

	return ioutil.ReadAll(jsonFile)
}
