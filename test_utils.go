package grpcwot

import (
	"encoding/json"
	"fmt"
	"github.com/emicklei/proto"
	"github.com/linksmart/thing-directory/wot"
	"io/ioutil"
	"os"
)

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

	b.constructMessagesNested()

	proto.Walk(definition,
		proto.WithRPC(b.HandleRPC))

	return b
}

func unmarshalExpectedFileForDataSchemas(expectedFile string) map[string]*wot.DataSchema {
	jsonFile, err := os.Open(expectedFile)
	if err != nil {
		fmt.Println(err)
		return map[string]*wot.DataSchema{}
	}
	defer func(jsonFile *os.File) {
		jsonFile.Close()
	}(jsonFile)

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return map[string]*wot.DataSchema{}
	}
	r := map[string]*wot.DataSchema{}
	err = json.Unmarshal(byteValue, &r)
	if err != nil {
		return map[string]*wot.DataSchema{}
	}
	return r
}
