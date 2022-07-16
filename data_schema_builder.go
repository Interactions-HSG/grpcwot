package grpcwot

import (
	"errors"
	"github.com/Interactions-HSG/grpcwot/pkg/protofmt"
	"github.com/emicklei/proto"
	"github.com/linksmart/thing-directory/wot"
	"strings"
)

type refMesTuple struct {
	pm string // Parent message where the field of type message is included
	t  string // Type of the field == name of the referenced message
	n  string // name of the field
}

type dataSchemaBuilder struct {
	ds map[string]*wot.DataSchema
	lm []refMesTuple
}

func newDataSchemaBuilder() *dataSchemaBuilder {
	return &dataSchemaBuilder{
		ds: map[string]*wot.DataSchema{},
		lm: []refMesTuple{},
	}
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
func (b *dataSchemaBuilder) HandleMessage(m *proto.Message) {
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
func (b *dataSchemaBuilder) fieldToDataSchema(f *proto.Field, messageName string) wot.DataSchema {
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

func (b *dataSchemaBuilder) oneofToDataSchema(oo *proto.Oneof, messageName string) []wot.DataSchema {
	oof := []wot.DataSchema{}
	for _, v := range oo.Elements {
		oof = append(oof, b.fieldToDataSchema(v.(*proto.OneOfField).Field, messageName))
	}
	return oof
}

func (b *dataSchemaBuilder) resolveSingleReference(elem refMesTuple) (string, error) {
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

// Resolves all references in lm by calling resolveSingleReference on them
// Sets the t (type) property of each refMesTuple to the full message name of the referenced message
func (b *dataSchemaBuilder) resolveAllReferences() error {
	for k, v := range b.lm {
		res, err := b.resolveSingleReference(v)
		if err != nil {
			return err
		}
		v.t = res
		b.lm[k] = v
	}
	return nil
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

// Extend the data schemas with references to other data schemas representing the message references
func (b *dataSchemaBuilder) constructMessagesNested() error {
	err := b.resolveAllReferences()
	if err != nil {
		return err
	}

	err = containsCircle(b.lm)
	if err != nil {
		return err
	}

	for _, v := range b.lm {
		b.ds[v.pm].ObjectSchema.Properties[v.n] = *b.ds[v.t]
	}
	return nil
}

// Walks the proto files messages and generates the data schemes
// In case of an invalid proto file an error is raised
func generateDataSchemas(protoFile *proto.Proto) (*dataSchemaBuilder, error) {
	b := newDataSchemaBuilder()

	proto.Walk(protoFile,
		proto.WithMessage(b.HandleMessage))

	err := b.constructMessagesNested()

	return b, err
}
