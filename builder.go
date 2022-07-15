package grpcwot

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/emicklei/proto"
	"github.com/linksmart/thing-directory/wot"
)

type builder struct {
	td   wot.ThingDescription
	dsb  *dataSchemaBuilder
	iab  *interactionAffordanceBuilder
	ip   string
	port int
	ac   map[string]affClassConfig
}

type affClassConfig struct {
	AffClass string
	Name     string `json:"Name,omitempty"`
}

func newBuilder(ip string, port int, dsb *dataSchemaBuilder) *builder {
	return &builder{
		td: wot.ThingDescription{
			Properties: map[string]wot.PropertyAffordance{},
			Actions:    map[string]wot.ActionAffordance{},
			Events:     map[string]wot.EventAffordance{},
		},
		dsb:  dsb,
		ip:   ip,
		port: port,
		ac:   map[string]affClassConfig{},
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

// getForms is a helper method to build the forms
func (b *builder) getForms(n string, ops []string) []wot.Form {
	return []wot.Form{
		{
			Href:        b.GetIRI(n),
			ContentType: "application/grpc+proto",
			Op:          ops,
		},
	}
}

// saveToAffClass is a helper function to save affordances in the affordance classification so they could be
// transformed into json and be reused for further builds
func (b *builder) saveToAffClass(k, n, affClass string) {
	if k == n {
		b.ac[k] = affClassConfig{
			AffClass: affClass,
		}
	} else {
		b.ac[k] = affClassConfig{
			Name:     n,
			AffClass: affClass,
		}
	}
}

// saveProperty converts and saves a RPC function to a Property Affordance in the TD
func (b *builder) saveProperty(p combinedProperties) {
	affordance := wot.PropertyAffordance{}
	var ops []string
	switch p.category {
	case 0:
		b.saveToAffClass(p.getProp.name, p.name, "property")
		affordance.DataSchema = *p.getProp.res
		ops = []string{"readproperty"}
	case 1:
		b.saveToAffClass(p.setProp.name, p.name, "property")
		affordance.DataSchema = *p.setProp.req
		ops = []string{"writeproperty"}
	case 2:
		b.saveToAffClass(p.getProp.name, p.name, "property")
		b.saveToAffClass(p.setProp.name, p.name, "property")
		affordance.DataSchema = *p.getProp.res
		ops = []string{"readproperty", "writeproperty"}
	default:
		return
	}

	affordance.Forms = b.getForms(p.name, ops)
	b.td.Properties[p.name] = affordance
}

// saveAction converts and saves a RPC function to an Action Affordance in the TD
func (b *builder) saveAction(r affs) {
	affordance := wot.ActionAffordance{}
	affordance.Input = *r.req
	affordance.Output = *r.res
	affordance.Forms = b.getForms(r.name, []string{})
	b.td.Actions[r.name] = affordance

	b.saveToAffClass(r.name, r.name, "action")
}

// saveEvent converts and saves a RPC function to an Event Affordance in the TD
func (b *builder) saveEvent(r affs) {
	affordance := wot.EventAffordance{}
	affordance.Data = *r.res
	affordance.Forms = b.getForms(r.name, []string{})
	b.td.Events[r.name] = affordance

	b.saveToAffClass(r.name, r.name, "event")
}

// readInput is a helper function to read in input from the user
func readInput(reader *bufio.Reader, s, k string, v []string) string {
	allowedInputs := []string{"", "a", "p", "e"}
	for {
		if len(v) == 1 {
			fmt.Printf("%s '%s' with RPC function '%s'\n", s, k, v[0])
		} else {
			fmt.Printf("%s '%s' with RPC functions '%s' and '%s'\n", s, k, v[0], v[1])
		}
		fmt.Print("->")
		t, _ := reader.ReadString('\n')
		t = strings.TrimSpace(t)
		if contains(allowedInputs, t) {
			return t
		}
	}
}

// categorizeAffordancesWithUserInput asks the user for validation of the made classification decisions and saves the
// affordances to the TD
func (b *builder) categorizeAffordancesWithUserInput() {
	fmt.Println("The following interaction affordances are already classified according to specific criterias." +
		"If you want to change the classification for a specific affordance please enter")
	fmt.Println("- (p) for property")
	fmt.Println("- (a) for action or")
	fmt.Println("- (e) for event")
	fmt.Println("If the classification is already correct, press enter.")

	reader := bufio.NewReader(os.Stdin)

	if len(b.iab.affC.combinedProp) != 0 {
		fmt.Println("The following were considered as properties: ")
	}
	for _, v := range b.iab.affC.combinedProp {
		t := readInput(reader, "Property", v.name, []string{v.getProp.name, v.setProp.name})
		switch t {
		case "":
			fallthrough
		case "p":
			b.saveProperty(v)
		case "a":
			switch v.category {
			case 2:
				fmt.Println("Should only the setter become an action (set) or both (both)?")
				fmt.Print("->")
				t, _ := reader.ReadString('\n')
				t = strings.TrimSpace(t)
				if t == "set" {
					b.saveAction(v.setProp)
					v.setProp = affs{}
					v.category = 0
					b.saveProperty(v)
				} else if t == "both" {
					b.saveAction(v.getProp)
					b.saveAction(v.setProp)
				}
			case 1:
				b.saveAction(v.setProp)
			case 0:
				b.saveAction(v.getProp)
			}
		case "e":
			if v.category == 2 {
				v.category = 1
				b.saveProperty(v)
			}
			b.saveEvent(v.getProp)
		}
	}
	if len(b.iab.affC.action) != 0 {
		fmt.Println("The following were considered as actions: ")
	}
	for _, v := range b.iab.affC.action {
		t := readInput(reader, "Action", v.name, []string{v.name})
		switch t {
		case "":
			fallthrough
		case "a":
			b.saveAction(v)
		case "p":
			if v.req.Type == "" {
				b.saveProperty(combinedProperties{
					name:     v.name,
					getProp:  v,
					category: 0,
				})
			} else {
				b.saveProperty(combinedProperties{
					name:     v.name,
					setProp:  v,
					category: 1,
				})
			}
		case "e":
			b.saveEvent(v)
		}
	}
	if len(b.iab.affC.event) != 0 {
		fmt.Println("The following were considered as events: ")
	}
	for _, v := range b.iab.affC.event {
		t := readInput(reader, "Event", v.name, []string{v.name})
		switch t {
		case "":
			fallthrough
		case "e":
			b.saveEvent(v)
		case "p":
			b.saveProperty(combinedProperties{
				name:     v.name,
				getProp:  v,
				category: 0,
			})
		case "a":
			b.saveAction(v)
		}
	}
}

//saveAfterConfigRPC Saves the affordances to the TD with the classification derived from the vonfig
func (b *builder) saveAfterConfigRPC() {
	for _, v := range b.iab.affC.combinedProp {
		b.saveProperty(v)
	}
	for _, v := range b.iab.affC.action {
		b.saveAction(v)
	}
	for _, v := range b.iab.affC.event {
		b.saveEvent(v)
	}
}

// contains, helper function do determine if a slice of type string contains the string s
func contains(a []string, s string) bool {
	for _, v := range a {
		if v == s {
			return true
		}
	}
	return false
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
		proto.WithService(b.HandleService))

	b.iab, err = generateInteractionAffordances(definition, dsb)

	if b.iab == nil {
		return err
	}

	b.categorizeAffordancesWithUserInput()

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
