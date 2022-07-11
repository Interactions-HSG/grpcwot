package grpcwot

import (
	"errors"
	"github.com/emicklei/proto"
	"github.com/linksmart/thing-directory/wot"
	"strings"
)

type interactionAffordanceBuilder struct {
	rpcs []*proto.RPC
	affs map[string]affs
	dsb  *dataSchemaBuilder
	affC affClasses
	cats catProps
}

type catProps struct {
	prop   checkCondition
	action checkCondition
	event  checkCondition
}

type affClasses struct {
	prop   []affs
	action []affs
	event  []affs
}

type affs struct {
	name string
	req  *wot.DataSchema
	res  *wot.DataSchema
}

func newInteractionAffordanceBuilder(dsb *dataSchemaBuilder) *interactionAffordanceBuilder {
	return &interactionAffordanceBuilder{
		[]*proto.RPC{},
		map[string]affs{},
		dsb,
		affClasses{
			[]affs{},
			[]affs{},
			[]affs{},
		},
		catProps{
			or(startsWithGetCaseInsensitive, startsWithSetCaseInsensitive),
			defaultConfig,
			and(not(hasRequestType), hasReturnType),
		},
	}
}

func (b *interactionAffordanceBuilder) HandleRPC(r *proto.RPC) {
	b.rpcs = append(b.rpcs, r)
}

func (b *interactionAffordanceBuilder) conformRPCs() error {
	b.affs = map[string]affs{}
	for _, v := range b.rpcs {
		if _, found := b.affs[v.Name]; found {
			return errors.New("Duplicate RPC name found in proto file for RPC name " + v.Name)
		}
		req, found := b.dsb.ds[v.RequestType]
		if !found {
			return errors.New("Not able to determine message for request type " + v.RequestType + " in RPC " + v.Name)
		}
		res, found := b.dsb.ds[v.ReturnsType]
		if !found {
			return errors.New("Not able to determine message for return type " + v.ReturnsType + " in RPC " + v.Name)
		}
		b.affs[v.Name] = affs{
			v.Name,
			req,
			res,
		}
	}
	return nil
}

type checkCondition func(affs) bool

func defaultConfig(_ affs) bool {
	return true
}

func startsWithGetCaseInsensitive(a affs) bool {
	return strings.HasPrefix(strings.ToUpper(a.name), "GET")
}

func startsWithSetCaseInsensitive(a affs) bool {
	return strings.HasPrefix(strings.ToUpper(a.name), "SET")
}

func startsWithGet(a affs) bool {
	return strings.HasPrefix(a.name, "Get")
}

func startsWithSet(a affs) bool {
	return strings.HasPrefix(a.name, "Set")
}

func typeNotEmpty(t *wot.DataSchema) bool {
	return t.ObjectSchema != nil &&
		t.Properties != nil &&
		len(t.Properties) != 0
}

func hasReturnType(a affs) bool {
	return typeNotEmpty(a.res)
}

func hasRequestType(a affs) bool {
	return typeNotEmpty(a.req)
}

func and(condition checkCondition, condition2 checkCondition) checkCondition {
	return func(a affs) bool {
		return condition(a) && condition2(a)
	}
}

func or(condition checkCondition, condition2 checkCondition) checkCondition {
	return func(a affs) bool {
		return condition(a) || condition2(a)
	}
}

func not(condition checkCondition) checkCondition {
	return func(a affs) bool {
		return !condition(a)
	}
}

func (b *interactionAffordanceBuilder) categorizeRPCs() {
	for _, v := range b.affs {
		switch {
		case b.cats.prop(v):
			b.affC.prop = append(b.affC.prop, v)
		case b.cats.event(v):
			b.affC.event = append(b.affC.event, v)
		default:
			b.affC.action = append(b.affC.action, v)
		}
	}
}

func generateInteractionAffordances(protoFile *proto.Proto, dsb *dataSchemaBuilder) (*interactionAffordanceBuilder, error) {
	b := newInteractionAffordanceBuilder(dsb)

	proto.Walk(protoFile,
		proto.WithRPC(b.HandleRPC))

	err := b.conformRPCs()
	if err != nil {
		return nil, err
	}

	b.categorizeRPCs()

	return b, nil
}
