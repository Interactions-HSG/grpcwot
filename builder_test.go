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
type expectedAffClass struct {
	Props   map[string][]string
	Actions map[string]string
	Events  map[string]string
}

// Message parsing tests
// Proto files as input files are provided under testFiles/testMessages/proto
// The desired structure of data schemas are provided under testFiles/testMessages/jsonLD

// The input file standardMessage provides only one message with fields of scalar value types
// Every scalar value type listed for proto3 (https://developers.google.com/protocol-buffers/docs/proto3#scalar) is included with on field
// The test expects, that all scalar value types are converted into the corresponding json value type according to the table provided above
func TestMessageParserStandardMessage(t *testing.T) {
	MessageParserTestHelper(t, "standardMessage")
}

// The input file contains two messages. The message LinkedTest and the message Test, which has a field name "linkedTest",
// which is of type "LinkedTest" and therefor refers to the Linked Test.
// The test expects, that the field linkedTest in Test is filled with the fields of message type LinkedTest
func TestMessageParserLinkedMessages(t *testing.T) {
	MessageParserTestHelper(t, "linkedMessage")
}

// This test has four messages with different scalar value type fields. One Message is the Empty message with no fields
// The test expects that all messages are converted into correct DataSchemas
func TestMessageParserMultipleMessages(t *testing.T) {
	MessageParserTestHelper(t, "multipleMessages")
}

// This test has two messages on the top level.
// The first message holds two nested messages inside. These nested messages are referenced as fields of the outer
// message and therefor the test expects the inner messages structure to be inserted into the corresponding fields.
// It also expects that the inner messages are included in the file by their complete reference name
// "<outerMessageName>.<innerMessageName"
// The second message tests if the behaviour still works if a second stage is introduced. So the outer message holds

func TestMessageParserNestedMessage(t *testing.T) {
	MessageParserTestHelper(t, "nestedMessage")
}

// This test adds another function to the input file from nestedMessages
// There should be tested if nested messages are also found and correctly inserted if they are referenced by other top
// level messages. For example BasicTest addresses a nested message InnerTest1, which is nested in OuterTest1
func TestMessageParserComplexNestedMessage(t *testing.T) {
	MessageParserTestHelper(t, "complexNestedMessage")
}

// Helper functions for message testing

func MessageParserTestHelper(t *testing.T, f string) {
	a := parseProtoFileToBuilder("testFiles/testMessages/protos/"+f+".proto", "").ds
	e := map[string]*wot.DataSchema{}
	unmarshalExpectedFile("testFiles/testMessages/jsonLD/"+f+".json", func(b []byte) error {
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

// Check automatic affordance classification
// Proto files as input files are provided under testFiles/testAffordanceClassification/proto
// The desired classification expressed in json format is stored in testFiles/testAffordanceClassification/jsonLD

// The input file defines different rpc functions.
// Functions are defined, which should be categorized as properties functions due to their name (get/set) and parameters
//		GetTest1 + SetTest1 => Test1	(normal get with empty request and set with data in response)
// 		GetTest2 + SetTest2 => Test2	(normal get with empty request and set with no data in response)
//		GetTest3			=> Test3 	(normal get with empty request and no set function provided)
//		SetTest4			=> Test4	(normal set without response data and no get function provided)
//		SetTest5			=> Test5	(normal set with response data and no get function provided)
// Functions are defined which should be categorized as actions as they are not properties by their name and not events,
// as they have request data, if they also have response data
//		ActionTest1 	(action with request and response data)
//		ActionTest2		(action with request and no response data)
//		ActionTest3		(action with no request data and no response data)
// Events are defined which should be categorized as events as they have no request, but response data
//		EventTest1		(event with no request, but response data)
func TestAutomaticAffordanceClassificationSimple(t *testing.T) {
	AffordanceClassificationTestHelper(t, "simple")
}

// Automatic affordance classification specific test helper functions

func AffordanceClassificationTestHelper(t *testing.T, f string) {
	a := parseProtoFileToBuilder("testFiles/testAffordanceClassification/protos/"+f+".proto", "").af
	e := expectedAffClass{}
	unmarshalExpectedFile("testFiles/testAffordanceClassification/jsonLD/"+f+".json", func(b []byte) error {
		return json.Unmarshal(b, &e)
	})
	checkAffordanceClassification(t, a, &e)
}

func checkAffordanceClassification(t *testing.T, a affClass, e *expectedAffClass) {
	checkProperties(t, a.props, e.Props)
	checkActionsOrEvents(t, a.actions, e.Actions, "Actions")
	checkActionsOrEvents(t, a.events, e.Events, "Events")
}

func checkActionsOrEvents(t *testing.T, a map[string]*proto.RPC, e map[string]string, d string) {
	if len(a) != len(e) {
		t.Errorf("The number of %s differ for expected and actual. "+
			"Number of expected is %d, number of actual is %d", d, len(e), len(a))
		printExpectedVersusActualActionsOrEvents(t, a, e, d)
		return
	}
	for k, v := range e {
		v2, ok := a[k]
		if !ok {
			t.Errorf("The following expected %s %s was not found in the actual %s.", d, k, d)
			printExpectedVersusActualActionsOrEvents(t, a, e, d)
			return
		}
		if v != v2.Name {
			t.Errorf("The following RPC function under the assigned name %s does not match. Expected was %s, but was %s", k, v, v2.Name)
		}
	}
}

func printExpectedVersusActualActionsOrEvents(t *testing.T, a map[string]*proto.RPC, e map[string]string, d string) {
	t.Errorf("Expected %s are:", d)
	for k, v := range e {
		t.Errorf("%s with RPCs: [%s]", k, v)
	}
	t.Errorf("Actual %s are:", d)
	for k, v := range a {
		t.Errorf("%s with RPCs: [%s]", k, v.Name)
	}
}

func checkProperties(t *testing.T, a map[string][]*proto.RPC, e map[string][]string) {
	if len(a) != len(e) {
		t.Errorf("The number of properties differ for expected and actual. "+
			"Number of expected is %d, number of actual is %d", len(e), len(a))
		printExpectedVersusActualProperties(t, a, e)
		return
	}
	for k, v := range e {
		v2, ok := a[k]
		if !ok {
			t.Errorf("The following expected property %s was not found in the actual properties.", k)
			printExpectedVersusActualProperties(t, a, e)
			return
		}
		if len(v) != len(v2) {
			t.Errorf("The assigned RPC function for the property %s don't match by length", k)
		}
		for i := 0; i < len(v); i++ {
			if v2[i].Name != v[i] {
				t.Errorf("The assigned RPC functions differ for property %s. Expected function is %s, actual function is %s", k, v[i], v2[i].Name)
				return
			}
		}
	}
}

func printExpectedVersusActualProperties(t *testing.T, a map[string][]*proto.RPC, e map[string][]string) {
	t.Errorf("Expected properties are:")
	for k, v := range e {
		switch len(v) {
		case 0:
		case 1:
			t.Errorf("Property: %s with RPCs: [%s]", k, v[0])
		case 2:
			t.Errorf("Property: %s with RPCs: [%s, %s]", k, v[0], v[1])
		}
	}
	t.Errorf("Actual properties are:")
	for k, v := range a {
		switch len(v) {
		case 0:
		case 1:
			t.Errorf("Property: %s with RPCs: [%s]", k, v[0].Name)
		case 2:
			t.Errorf("Property: %s with RPCs: [%s, %s]", k, v[0].Name, v[1].Name)
		}
	}
}

// Configuration file based classification tester

// Check configuration file based affordance classification
// Proto files as input files are provided under testFiles/testFileBasedClassification/proto
// Config files in json format as input files are provided under testFiles/testFileBasedClassification/configFiles
// The desired classification expressed in json format is stored in testFiles/testFileBasedClassification/jsonLD

// The input file holds 6 rpc functions
//		GetTest1 and SetTest1 are defined in the config file to be concluded to the property Test1
// 		GetTest2 and SetTest2 would be classified as property Test2, when automatic classification would be applied, but
//			SetTest2 is defined in the config file to form an action named SetTest2Action and GetTest2 is defined to form
//			a property with predefined name Test2Action
//		ActionTest1 is defined to be action with same name as automatically classified
//		EventTest1 is defined to be event with same name as automatically classified
func TestFileBasedAffordanceClassificationSimple(t *testing.T) {
	FileBasedClassificationTestHelper(t, "simple")
}

func FileBasedClassificationTestHelper(t *testing.T, f string) {
	baseUrl := "testFiles/testFileBasedClassification/"
	a := parseProtoFileToBuilder(
		baseUrl+"protos/"+f+".proto",
		baseUrl+"configFiles/"+f+".json",
	)
	e := expectedAffClass{}
	unmarshalExpectedFile(baseUrl+"jsonLD/"+f+".json", func(b []byte) error {
		return json.Unmarshal(b, &e)
	})
	checkAffordanceClassification(t, a.af, &e)
}

// General helper functions

func parseProtoFileToBuilder(protoFile string, configFile string) (b *builder) {
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

	b.constructMessagesNested()

	if configFile != "" {
		// For tests running with config file
		byteValue, _ := readByteValueFromJsonFile(configFile)
		err = json.Unmarshal(byteValue, &b.ac)
		if err != nil {
			return nil
		}

		proto.Walk(definition,
			proto.WithRPC(b.HandleRPCWithConfig))
	} else {
		// For tests running without config file
		proto.Walk(definition,
			proto.WithRPC(b.HandleRPC))
	}

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
