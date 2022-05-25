package grpcwot

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Interactions-HSG/grpcwot/pkg/protofmt"
	"github.com/emicklei/proto"
	"github.com/linksmart/thing-directory/wot"
)

type builder struct {
	td   wot.ThingDescription
	ds   map[string]*wot.DataSchema
	ip   string
	port int
}

func newBuilder(ip string, port int) *builder {
	return &builder{
		td:   wot.ThingDescription{},
		ds:   map[string]*wot.DataSchema{},
		ip:   ip,
		port: port,
	}
}

// GetIRI returns the target IRI for the RPC
func (b *builder) GetIRI(rpcName string) string {
	return fmt.Sprintf("http://%s:%d/%s/%s", b.ip, b.port, b.td.Title, rpcName)
}

// HandleService assigns the Title for the resulting TD
func (b *builder) HandleService(s *proto.Service) {
	b.td.Title = s.Name
}

// HandleRPC constructs Interaction Affordances with prefetched DataSchema
func (b *builder) HandleRPC(r *proto.RPC) {
	// TODO: affordance classification should happen here
	// putting all RPC to Actions for now
	if b.td.Actions == nil {
		b.td.Actions = map[string]wot.ActionAffordance{}
	}
	affordance := wot.ActionAffordance{}
	affordance.Input = *b.ds[r.RequestType]
	affordance.Output = *b.ds[r.ReturnsType]
	affordance.Forms = []wot.Form{
		{
			Href:        b.GetIRI(r.Name),
			ContentType: "application/grpc+proto",
			Op:          []string{"writeproperty"},
		},
	}

	b.td.Actions[r.Name] = affordance
}

// HandleMessage build a DataSchema: https://www.w3.org/TR/wot-thing-description/#dataschema
// from a Message in the protobuf definition
func (b *builder) HandleMessage(m *proto.Message) {
	if _, ok := b.ds[m.Name]; !ok {
		b.ds[m.Name] = &wot.DataSchema{
			DataType: "object",
			ObjectSchema: &wot.ObjectSchema{
				Properties: map[string]wot.DataSchema{},
			},
		}
	}
	for _, v := range m.Elements {
		switch protofmt.NameOfVisitee(v) {
		case "NormalField":
			b.ds[m.Name].ObjectSchema.Properties[v.(*proto.NormalField).Field.Name] =
				fieldToDataSchema(v.(*proto.NormalField).Field)
		case "Comment":
		case "Oneof":
			b.ds[m.Name].ObjectSchema.Properties[v.(*proto.Oneof).Name] =
				wot.DataSchema{OneOf: oneofToDataSchema(v.(*proto.Oneof))}
		}
	}
}

func fieldToDataSchema(f *proto.Field) wot.DataSchema {
	return wot.DataSchema{DataType: determineJsonTypeForField(f)}
}

// determineJsonTypeForField is a helper function to convert datatypes used in proto files into data types used in
// json format.
func determineJsonTypeForField(f *proto.Field) string {
	switch f.Type {
	case "double", "float":
		return "number"
	case "int32", "int64", "uint32", "uint64", "sint32", "sint64", "fixed32", "fixed64", "sfixed32", "sfixed64":
		return "integer"
	case "bool":
		return "boolean"
	case "string", "bytes":
		return "string"
	default:
		return "object"
	}
}

func oneofToDataSchema(oo *proto.Oneof) []wot.DataSchema {
	oof := []wot.DataSchema{}
	for _, v := range oo.Elements {
		oof = append(oof, fieldToDataSchema(v.(*proto.OneOfField).Field))
	}
	return oof
}

// GenerateTDfromProtoBuf parses `protoFile` to generate `tdFile`
func GenerateTDfromProtoBuf(protoFile, tdFile, ip string, port int) error {
	// initialize the TD builder with an empty TD and DataSchema
	b := newBuilder(ip, port)

	// parse the protoFile with the emicklei/proto
	reader, _ := os.Open(protoFile)
	defer reader.Close()
	parser := proto.NewParser(reader)
	definition, err := parser.Parse()
	if err != nil {
		return nil
	}

	// extract the Messages as DataSchema
	proto.Walk(definition,
		proto.WithService(b.HandleService),
		proto.WithMessage(b.HandleMessage))

	// translate the RPC functions into Interaction Affordances
	proto.Walk(definition,
		proto.WithRPC(b.HandleRPC))

	// serialize the TD to JSONLD
	tdBytes, _ := json.Marshal(b.td)
	f, err := os.Create(tdFile)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(tdBytes)
	if err != nil {
		return err
	}
	return nil
}
