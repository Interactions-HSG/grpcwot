package grpcwot

import (
	"encoding/json"
	"fmt"
	"github.com/emicklei/proto"
	"io/ioutil"
	"os"
	"testing"
)

type expectedAffClass struct {
	Props   map[string][]string
	Actions map[string]string
	Events  map[string]string
}

func newEmptyExpectedAffClass() *expectedAffClass {
	return &expectedAffClass{
		Props:   map[string][]string{},
		Actions: map[string]string{},
		Events:  map[string]string{},
	}
}

func TestAutomaticAffordanceClassificationSimple(t *testing.T) {
	AffordanceClassificationTestHelper(t, "simple")
}

func AffordanceClassificationTestHelper(t *testing.T, f string) {
	a := parseProtoFileToBuilder("testFiles/testAffordanceClassification/protos/" + f + ".proto").af
	e := unmarshalExpectedFileForAffClass("testFiles/testAffordanceClassification/jsonLD/" + f + ".json")
	checkAffordanceClassification(t, a, e)
}

func unmarshalExpectedFileForAffClass(expectedFile string) *expectedAffClass {
	jsonFile, err := os.Open(expectedFile)
	if err != nil {
		fmt.Println(err)
		return newEmptyExpectedAffClass()
	}
	defer func(jsonFile *os.File) {
		jsonFile.Close()
	}(jsonFile)

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return newEmptyExpectedAffClass()
	}
	r := expectedAffClass{}
	err = json.Unmarshal(byteValue, &r)
	if err != nil {
		return newEmptyExpectedAffClass()
	}
	return &r
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
