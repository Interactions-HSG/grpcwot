package grpcwot

import (
	"encoding/json"
	"github.com/emicklei/proto"
	"os"
	"testing"
)

type expectedAffClass_ struct {
	Props   map[string][]string
	Actions map[string]string
	Events  map[string]string
}

func newEmptyExpectedAffClass_() *expectedAffClass {
	return &expectedAffClass{
		Props:   map[string][]string{},
		Actions: map[string]string{},
		Events:  map[string]string{},
	}
}

func TestFileBasedAffordanceClassificationSimple(t *testing.T) {
	FileBasedClassificationTestHelper(t, "simple")
}

func FileBasedClassificationTestHelper(t *testing.T, f string) {
	baseUrl := "testFiles/testFileBasedClassification/"
	a := parseProtoFileWithConfigToBuilder(
		baseUrl+"protos/"+f+".proto",
		baseUrl+"configFiles/"+f+".json",
	)
	e := unmarshalExpectedFileForAffClass(baseUrl + "jsonLD/" + f + ".json")
	checkAffordanceClassification(t, a.af, e)
}

func parseProtoFileWithConfigToBuilder(protoFile string, configFile string) (b *builder) {
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

	byteValue, _ := readByteValueFromJsonFile(configFile)
	err = json.Unmarshal(byteValue, &b.ac)

	proto.Walk(definition,
		proto.WithRPC(b.HandleRPCWithConfig))

	return b
}
