package grpcwot

import (
	"errors"
	"github.com/emicklei/proto"
	"github.com/linksmart/thing-directory/wot"
)

type interactionAffordanceBuilder struct {
	rpcs []*proto.RPC
	affs map[string]affs
	dsb  *dataSchemaBuilder
}

type affs struct {
	req *wot.DataSchema
	res *wot.DataSchema
}

func newInteractionAffordanceBuilder(dsb *dataSchemaBuilder) *interactionAffordanceBuilder {
	return &interactionAffordanceBuilder{
		[]*proto.RPC{},
		map[string]affs{},
		dsb,
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
			req,
			res,
		}
	}
	return nil
}

func (b *interactionAffordanceBuilder) prefilterRPCs() {

}

func generateInteractionAffordances(protoFile *proto.Proto, dsb *dataSchemaBuilder) (*interactionAffordanceBuilder, error) {
	b := newInteractionAffordanceBuilder(dsb)

	proto.Walk(protoFile,
		proto.WithRPC(b.HandleRPC))

	err := b.conformRPCs()
	if err != nil {
		return nil, err
	}

	return b, nil
}
