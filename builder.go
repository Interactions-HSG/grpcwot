package grpcwot

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/emicklei/proto"
	"github.com/linksmart/thing-directory/wot"
)

type builder struct {
	td   wot.ThingDescription
	dsb  *dataSchemaBuilder
	ip   string
	port int
}

func newBuilder(ip string, port int, dsb *dataSchemaBuilder) *builder {
	return &builder{
		td:   wot.ThingDescription{},
		dsb:  dsb,
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
	affordance.Input = *b.dsb.ds[r.RequestType]
	affordance.Output = *b.dsb.ds[r.ReturnsType]
	affordance.Forms = []wot.Form{
		{
			Href:        b.GetIRI(r.Name),
			ContentType: "application/grpc+proto",
			Op:          []string{"writeproperty"},
		},
	}

	b.td.Actions[r.Name] = affordance
}

// GenerateTDfromProtoBuf parses `protoFile` to generate `tdFile`
func GenerateTDfromProtoBuf(protoFile, tdFile, ip string, port int) error { // parse the protoFile with the emicklei/proto
	reader, _ := os.Open(protoFile)
	defer reader.Close()
	parser := proto.NewParser(reader)
	definition, err := parser.Parse()
	if err != nil {
		return nil
	}

	// Read the Messages and produce DataSchemes
	dsb, err := generateDataSchemas(definition)
	if err != nil {
		return err
	}

	// initialize the TD builder with an empty TD and DataSchema
	b := newBuilder(ip, port, dsb)

	// translate the RPC functions into Interaction Affordances
	proto.Walk(definition,
		proto.WithService(b.HandleService),
		proto.WithRPC(b.HandleRPC))

	iab, err := generateInteractionAffordances(definition, dsb)

	if iab == nil {
		return err
	}

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
