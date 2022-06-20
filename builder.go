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

func getFullMessageName(m *proto.Message) string {
	if v, ok := m.Parent.(*proto.Message); ok {
		return getFullMessageName(v) + "." + m.Name
	} else {
		return m.Name
	}
}

// HandleMessage build a DataSchema: https://www.w3.org/TR/wot-thing-description/#dataschema
// from a Message in the protobuf definition
func (b *builder) HandleMessage(m *proto.Message) {
	fullMessageName := getFullMessageName(m)
	if _, ok := b.ds[fullMessageName]; !ok {
		b.ds[fullMessageName] = &wot.DataSchema{
			DataType: "object",
			ObjectSchema: &wot.ObjectSchema{
				Properties: map[string]wot.DataSchema{},
			},
		}
	}
	for _, v := range m.Elements {
		switch protofmt.NameOfVisitee(v) {
		case "NormalField":
			b.ds[fullMessageName].ObjectSchema.Properties[v.(*proto.NormalField).Field.Name] =
				fieldToDataSchema(v.(*proto.NormalField).Field)
		case "Comment":
		case "Oneof":
			b.ds[fullMessageName].ObjectSchema.Properties[v.(*proto.Oneof).Name] =
				wot.DataSchema{OneOf: oneofToDataSchema(v.(*proto.Oneof))}
		}
	}
}

// fieldToDataSchema converts the given proto's message field into a WoT DataScheme
// cf. https://www.w3.org/TR/wot-thing-description/#dataschema
func fieldToDataSchema(f *proto.Field) wot.DataSchema {
	var fieldType string
	switch f.Type {
	case "double", "float":
		fieldType = "number"
	case "int32", "int64", "uint32", "uint64", "sint32", "sint64", "fixed32", "fixed64", "sfixed32", "sfixed64":
		fieldType = "integer"
	case "bool":
		fieldType = "boolean"
	case "string", "bytes":
		fieldType = "string"
	default:
		fieldType = "object"
	}

	return wot.DataSchema{DataType: fieldType}
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
