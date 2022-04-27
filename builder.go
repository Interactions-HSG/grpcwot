package grpcwot

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Interactions-HSG/grpcwot/pkg/protofmt"
	"github.com/emicklei/proto"
	"github.com/linksmart/thing-directory/wot"
)

type tuple struct {
	pm string
	t  string
	n  string
}

type builder struct {
	td   wot.ThingDescription
	ds   map[string]*wot.DataSchema
	ip   string
	port int
	lm   []tuple
}

func newBuilder(ip string, port int) *builder {
	return &builder{
		td: wot.ThingDescription{
			Properties: map[string]wot.PropertyAffordance{},
			Actions:    map[string]wot.ActionAffordance{},
			Events:     map[string]wot.EventAffordance{},
		},
		ds:   map[string]*wot.DataSchema{},
		ip:   ip,
		port: port,
		lm:   []tuple{},
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

// contains, helper function do termine if a slice of type string a contains the string s
func contains(a []string, s string) bool {
	for _, v := range a {
		if v == s {
			return true
		}
	}
	return false
}

// HandleRPC is called on every RPC function and applies standard filters to build an initial classification of affordances
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
	mN := m.Name
	p := m.Parent
	for true {
		v, ok := p.(*proto.Message)
		if !ok {
			break
		}
		mN = v.Name + "." + mN
		p = v.Parent
	}
	if _, ok := b.ds[mN]; !ok {
		if len(m.Elements) != 0 {
			b.ds[mN] = &wot.DataSchema{
				DataType: "object",
				ObjectSchema: &wot.ObjectSchema{
					Properties: map[string]wot.DataSchema{},
				},
			}
		} else {
			b.ds[mN] = &wot.DataSchema{}
		}
	}
	for _, v := range m.Elements {
		switch protofmt.NameOfVisitee(v) {
		case "NormalField":
			b.ds[mN].ObjectSchema.Properties[v.(*proto.NormalField).Field.Name] =
				b.fieldToDataSchema(v.(*proto.NormalField).Field, mN)
		case "Comment":
		case "Oneof":
			b.ds[mN].ObjectSchema.Properties[v.(*proto.Oneof).Name] =
				wot.DataSchema{OneOf: b.oneofToDataSchema(v.(*proto.Oneof), mN)}
		}
	}
}

func (b *builder) fieldToDataSchema(f *proto.Field, pm string) wot.DataSchema {
	return wot.DataSchema{DataType: b.fieldToDataSchemaHelper(f, pm)}
}

// fieldToDataSchemaHelper is a helper method to convert datatypes used in proto files into data types used in
// json format. References to other messages are stored in lm to be evaluated later
func (b *builder) fieldToDataSchemaHelper(f *proto.Field, pm string) string {
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
		b.lm = append(b.lm, tuple{pm, f.Type, f.Name})
		return "object"
	}
}

func (b *builder) oneofToDataSchema(oo *proto.Oneof, pm string) []wot.DataSchema {
	oof := []wot.DataSchema{}
	for _, v := range oo.Elements {
		oof = append(oof, b.fieldToDataSchema(v.(*proto.OneOfField).Field, pm))
	}
	return oof
}

// isParentMessageOf is a helper function to determine if the actual message with name s is parent of any other nested
// message or could be filled into their parent messages
func isParentMessageOf(s string, lm []tuple) bool {
	for _, v := range lm {
		if s == v.pm {
			return true
		}
	}
	return false
}

// containsFalse is a helper function to check if all values in a bool splice are false
func containsFalse(a []bool) bool {
	for _, v := range a {
		if !v {
			return true
		}
	}
	return false
}

// constructMessagesNested tries to combine and build up nested trees by filling in the referenced messages into the
// parent messages
func (b *builder) constructMessagesNested() {
	a := make([]bool, len(b.lm), len(b.lm))
	for containsFalse(a) {
		for k, v := range b.lm {
			if a[k] || isParentMessageOf(v.n, b.lm) {
				continue
			}
			a[k] = true
			s := strings.Split(v.pm, ".")
			c := false
			for i := 0; i <= len(s); i++ {
				if _, ok := b.ds[strings.Join(s[i:], ".")+"."+v.t]; ok {
					b.ds[v.pm].ObjectSchema.Properties[v.n] = *b.ds[strings.Join(s[i:], ".")+"."+v.t]
					c = true
					break
				}
			}
			if !c {
				b.ds[v.pm].ObjectSchema.Properties[v.n] = *b.ds[v.t]
			}
		}
	}
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

	b.constructMessagesNested()

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
