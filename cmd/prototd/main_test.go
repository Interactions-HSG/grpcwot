package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Interactions-HSG/grpcwot"
)

// TestProtoToTD runs over the test proto files in ./test/*/input.proto and compare the result
// with output.jsonld in the same directory
func TestProtoToTD(t *testing.T) {
	testDir := "./test"
	tests, err := ioutil.ReadDir(testDir)
	if err != nil {
		t.Error(err)
	}
	for _, f := range tests {
		inputFile := filepath.Join(testDir, f.Name(), "input.proto")
		outputFile := filepath.Join(testDir, f.Name(), "output.jsonld")
		tmpFile, err := ioutil.TempFile(testDir, "result.jsonld")
		if err != nil {
			t.Error(err)
		}
		defer os.Remove(tmpFile.Name())
		grpcwot.GenerateTDfromProtoBuf(inputFile, tmpFile.Name(), "127.0.0.1", 50051)
		result, err := ioutil.ReadFile(tmpFile.Name())
		if err != nil {
			t.Error(err)
		}
		out, err := ioutil.ReadFile(outputFile)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(result, out) {
			t.Errorf("%v => \n%v, want \n%v", inputFile, result, out)
		}
	}
}
