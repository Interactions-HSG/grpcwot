package grpcwot

import (
	"fmt"
	"github.com/emicklei/proto"
	"io/ioutil"
	"os"
)

type unmarshalExpected func([]byte) error

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
