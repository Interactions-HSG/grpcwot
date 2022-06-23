package grpcwot

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Interactions-HSG/grpcwot/pkg/protofmt"
	"github.com/emicklei/proto"
	"github.com/linksmart/thing-directory/wot"
)

type refMesTuple struct {
	pm string // Parent message where the field of type message is included
	t  string // Type of the field == name of the referenced message
	n  string // Name of the field
}

type builder struct {
	td   wot.ThingDescription
	ds   map[string]*wot.DataSchema
	ip   string
	port int
	lm   []refMesTuple
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

// getFullMessageName returns the complete Message name from the root level to the actual message
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
				b.fieldToDataSchema(v.(*proto.NormalField).Field, fullMessageName)
		case "Comment":
		case "Oneof":
			b.ds[fullMessageName].ObjectSchema.Properties[v.(*proto.Oneof).Name] =
				wot.DataSchema{OneOf: b.oneofToDataSchema(v.(*proto.Oneof), fullMessageName)}
		}
	}
}

// fieldToDataSchema converts the given proto's message field into a WoT DataScheme
// cf. https://www.w3.org/TR/wot-thing-description/#dataschema
func (b *builder) fieldToDataSchema(f *proto.Field, messageName string) wot.DataSchema {
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
		b.lm = append(b.lm, refMesTuple{pm: messageName, t: f.Type, n: f.Name})
	}
	return wot.DataSchema{DataType: fieldType}
}

func (b *builder) oneofToDataSchema(oo *proto.Oneof, messageName string) []wot.DataSchema {
	oof := []wot.DataSchema{}
	for _, v := range oo.Elements {
		oof = append(oof, b.fieldToDataSchema(v.(*proto.OneOfField).Field, messageName))
	}
	return oof
}

func (b *builder) resolveSingleReference(elem refMesTuple) (string, error) {
	parts := strings.Split(elem.pm, ".")
	for k := len(parts); k >= 0; k-- {
		s := strings.Join(parts[:k], ".") + "." + elem.t
		if k == 0 {
			s = elem.t
		}
		if _, ok := b.ds[s]; ok {
			return s, nil
		}
	}
	return "", errors.New("No corresponding message found for type reference " + elem.t +
		" in message " + elem.pm)
}

// Determines if the parent Message is included in another message intrusion
func (b *builder) isParentMessage(elem refMesTuple) bool {
	referencedMessage, err := b.resolveSingleReference(elem)
	if err != nil {
		return false
	}
	for _, v := range b.lm {
		if v.pm == referencedMessage {
			return true
		}
	}
	return false
}

// containsCircle determines if the message references hold a circle and would lead to infinite injection
// For the search algorithm DFS would be an alternative with a better worst case performance
// but the structure of lm would have to be modified and for proto file's messages the average use case does not draw
// on long paths through the graph (the strengths of DFS)
func containsCircle(lm []refMesTuple) error {
	ended := make([]bool, len(lm))
	count := len(lm)
	last := count
	for count != 0 {
	out:
		for k, v := range lm {
			if ended[k] {
				continue
			}
			for k2, v2 := range lm {
				if ended[k2] {
					continue
				}
				if v.t == v2.pm {
					continue out
				}
			}
			ended[k] = true
			count--
		}
		if last == count {
			return errors.New("proto file contained circle reference in the Messages")
		}
	}
	return nil
}

// Resolves all message references stored in lm
func (b *builder) constructMessagesNested() error {
	for k, v := range b.lm {
		res, err := b.resolveSingleReference(v)
		if err != nil {
			return err
		}
		v.t = res
		b.lm[k] = v
	}

	err := containsCircle(b.lm)
	if err != nil {
		return err
	}

	var left []refMesTuple
	for len(b.lm) != 0 {
		for _, v := range b.lm {
			if b.isParentMessage(v) {
				left = append(left, v)
				continue
			}
			referencedMessage, err := b.resolveSingleReference(v)
			if err != nil {
				return err
			}
			b.ds[v.pm].ObjectSchema.Properties[v.n] = *b.ds[referencedMessage]
		}
		b.lm = left
		left = []refMesTuple{}
	}
	return nil
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
