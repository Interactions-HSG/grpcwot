package grpcwot

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Interactions-HSG/grpcwot/pkg/protofmt"
	"github.com/emicklei/proto"
	"github.com/linksmart/thing-directory/wot"
)

type affClass struct {
	props   map[string][]*proto.RPC
	actions map[string]*proto.RPC
	events  map[string]*proto.RPC
}

type affClassConfig struct {
	AffClass string
	Name     string `json:"Name,omitempty"`
}

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
	af   affClass
	ac   map[string]affClassConfig
	re   []string
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
		af: affClass{
			props:   map[string][]*proto.RPC{},
			actions: map[string]*proto.RPC{},
			events:  map[string]*proto.RPC{},
		},
		ac: map[string]affClassConfig{},
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

// isGetPropertyRPC is a function to define properties for classification as Get Property function
func (b *builder) isGetPropertyRPC(r *proto.RPC) bool {
	return strings.HasPrefix(r.Name, "Get") &&
		b.ds[r.RequestType].DataType == ""
}

// isSetPropertyRPC is a function to define properties for classification as Set Property function
func (b *builder) isSetPropertyRPC(r *proto.RPC) bool {
	return strings.HasPrefix(r.Name, "Set") &&
		b.ds[r.RequestType].DataType != ""
}

// isEventRPC is a function to define properties for classification as Event function
func (b *builder) isEventRPC(r *proto.RPC) bool {
	return b.ds[r.RequestType].DataType == "" &&
		b.ds[r.ReturnsType].DataType != ""
}

// setFitsGetProp is a helper function to check if get and set property function match
func setFitsGetProp(g, s *proto.RPC) bool {
	return g.ReturnsType == s.RequestType
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

// trimGetSet helper function to extract the property's name from Getter or Setter
func trimGetSet(n string) string {
	return strings.TrimLeft(strings.TrimLeft(n, "Set"), "Get")
}

// HandleRPC is called on every RPC function and applies standard filters to build an initial classification of affordances
func (b *builder) HandleRPC(r *proto.RPC) {
	if b.re != nil && !contains(b.re, r.Name) {
		return
	}
	if b.isGetPropertyRPC(r) {
		if v, ok := b.af.props[trimGetSet(r.Name)]; ok && len(v) == 1 {
			if setFitsGetProp(r, v[0]) {
				b.af.props[trimGetSet(r.Name)] = append(b.af.props[trimGetSet(r.Name)], r)
			} else {
				b.af.actions[v[0].Name] = v[0]
				b.af.props[trimGetSet(r.Name)] = []*proto.RPC{r}
			}
		} else {
			b.af.props[trimGetSet(r.Name)] = []*proto.RPC{r}
		}
	} else if b.isSetPropertyRPC(r) {
		if v, ok := b.af.props[trimGetSet(r.Name)]; ok && len(v) == 1 {
			if setFitsGetProp(v[0], r) {
				b.af.props[trimGetSet(r.Name)] = append(b.af.props[trimGetSet(r.Name)], r)
			} else {
				b.af.actions[r.Name] = r
			}
		} else {
			b.af.props[trimGetSet(r.Name)] = []*proto.RPC{r}
		}
	} else if b.isEventRPC(r) {
		b.af.events[r.Name] = r
	} else {
		b.af.actions[r.Name] = r
	}
}

// HandleRPCWithConfig classifies RPC functions to interaction affordances based on a provided configuration
func (b *builder) HandleRPCWithConfig(r *proto.RPC) {
	if b.re != nil && !contains(b.re, r.Name) {
		return
	}
	c, ok := b.ac[r.Name]
	if !ok {
		return
	}
	switch c.AffClass {
	case "property":
		b.af.props[c.Name] = append(b.af.props[c.Name], r)
	case "action":
		b.af.actions[c.Name] = r
	case "event":
		b.af.events[c.Name] = r
	}
}

// HandleMessage build a DataSchema: https://www.w3.org/TR/wot-thing-description/#dataschema
// from a Message in the protobuf definition
func (b *builder) HandleMessage(m *proto.Message) {
	mN := m.Name
	p := m.Parent
	for {
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
	var oof []wot.DataSchema
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
	a := make([]bool, len(b.lm))
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

// saveProperty converts and saves a RPC function to a Property Affordance in the TD
func (b *builder) saveProperty(n string, rs []*proto.RPC) {
	affordance := wot.PropertyAffordance{}

	ops := make([]string, len(rs))
	for i := 0; i < len(rs); i++ {
		b.saveToAffClass(rs[i].Name, n, "property")
		if b.ds[rs[i].RequestType].DataType == "" {
			ops[i] = "readproperty"
			affordance.DataSchema = *b.ds[rs[i].ReturnsType]
		} else {
			ops[i] = "writeproperty"
			affordance.DataSchema = *b.ds[rs[i].RequestType]
		}
	}

	affordance.Forms = b.getForms(n, ops)
	b.td.Properties[n] = affordance
}

// saveAction converts and saves a RPC function to an Action Affordance in the TD
func (b *builder) saveAction(n string, r *proto.RPC) {
	affordance := wot.ActionAffordance{}
	affordance.Input = *b.ds[r.RequestType]
	affordance.Output = *b.ds[r.ReturnsType]
	affordance.Forms = b.getForms(n, []string{"writeproperty"})
	b.td.Actions[n] = affordance

	b.saveToAffClass(n, r.Name, "action")
}

// saveEvent converts and saves a RPC function to an Event Affordance in the TD
func (b *builder) saveEvent(n string, r *proto.RPC) {
	affordance := wot.EventAffordance{}
	affordance.Data = *b.ds[r.ReturnsType]
	affordance.Forms = b.getForms(n, []string{"readproperty"})
	b.td.Events[n] = affordance

	b.saveToAffClass(n, r.Name, "event")
}

// readInput is a helper function to read in input from the user
func readInput(reader *bufio.Reader, s, k string, v []*proto.RPC) string {
	allowedInputs := []string{"", "a", "p", "e"}
	for {
		if len(v) == 1 {
			fmt.Printf("%s '%s' with RPC function '%s'\n", s, k, v[0].Name)
		} else {
			fmt.Printf("%s '%s' with RPC functions '%s' and '%s'\n", s, k, v[0].Name, v[1].Name)
		}
		fmt.Print("->")
		t, _ := reader.ReadString('\n')
		t = strings.TrimSpace(t)
		if contains(allowedInputs, t) {
			return t
		}
	}
}

// checkAndSaveInteractionAffordances asks the user for validation of the made classification decisions and saves the
// affordances to the TD
func (b *builder) checkAndSaveInteractionAffordances() {
	fmt.Println("The following interaction affordances are already classified according to specific criterias." +
		"If you want to change the classification for a specific affordance please enter")
	fmt.Println("- (p) for property")
	fmt.Println("- (a) for action or")
	fmt.Println("- (e) for event")
	fmt.Println("If the classification is already correct, press enter.")

	reader := bufio.NewReader(os.Stdin)

	if len(b.af.props) != 0 {
		fmt.Println("The following were considered as properties: ")
	}
	for k, v := range b.af.props {
		t := readInput(reader, "Property", k, v)
		switch t {
		case "":
			fallthrough
		case "p":
			b.saveProperty(k, v)
		case "a":
			if len(v) == 2 {
				fmt.Println("Should only the setter become an action (set) or both (both)?")
				fmt.Print("->")
				t, _ := reader.ReadString('\n')
				t = strings.TrimSpace(t)
				if t == "set" {
					if b.ds[v[0].RequestType].DataType == "" {
						b.saveProperty(k, []*proto.RPC{v[0]})
						b.saveAction(v[1].Name, v[1])
					} else {
						b.saveProperty(k, []*proto.RPC{v[1]})
						b.saveAction(v[0].Name, v[0])
					}
				} else if t == "both" {
					b.saveAction(v[0].Name, v[0])
					b.saveAction(v[1].Name, v[1])
				}
			} else {
				b.saveAction(v[0].Name, v[0])
			}
		case "e":
			if len(v) == 2 {
				if b.ds[v[0].RequestType].DataType != "" {
					b.saveProperty(k, []*proto.RPC{v[0]})
					b.saveEvent(v[1].Name, v[1])
				} else {
					b.saveProperty(k, []*proto.RPC{v[1]})
					b.saveEvent(v[0].Name, v[0])
				}
			} else {
				b.saveEvent(v[0].Name, v[0])
			}
		}
	}
	if len(b.af.actions) != 0 {
		fmt.Println("The following were considered as actions: ")
	}
	for k, v := range b.af.actions {
		t := readInput(reader, "Action", k, []*proto.RPC{v})
		switch t {
		case "":
			fallthrough
		case "a":
			b.saveAction(k, v)
		case "p":
			b.saveProperty(trimGetSet(k), []*proto.RPC{v})
		case "e":
			b.saveEvent(k, v)
		}
	}
	if len(b.af.events) != 0 {
		fmt.Println("The following were considered as events: ")
	}
	for k, v := range b.af.events {
		t := readInput(reader, "Event", k, []*proto.RPC{v})
		switch t {
		case "":
			fallthrough
		case "e":
			b.saveEvent(k, v)
		case "p":
			b.saveProperty(trimGetSet(k), []*proto.RPC{v})
		case "a":
			b.saveAction(k, v)
		}
	}
}

//saveAfterConfigRPC Saves the affordances to the TD with the classification derived from the vonfig
func (b *builder) saveAfterConfigRPC() {
	for k, v := range b.af.props {
		b.saveProperty(k, v)
	}
	for k, v := range b.af.actions {
		b.saveAction(k, v)
	}
	for k, v := range b.af.events {
		b.saveEvent(k, v)
	}
}

// Generates a json file to store the configurations made by classification process
func (b *builder) generateConfigFileForAffordanceClassification(configFile string) {
	configBytes, _ := json.Marshal(b.ac)
	f, err := os.Create(configFile)
	if err != nil {
		return
	}
	defer f.Close()
	_, err = f.Write(configBytes)
	if err != nil {
		return
	}
}

// GenerateTDfromProtoBuf parses `protoFile` to generate `tdFile`
func GenerateTDfromProtoBuf(protoFile, outputDir, classConfigFile, reducedConfigFile, ip string, port int) error {
	configSet := true
	// Check if config File is present
	if _, err := os.Stat(classConfigFile); errors.Is(err, os.ErrNotExist) {
		configSet = false
	}

	// initialize the TD builder with an empty TD and DataSchema
	b := newBuilder(ip, port)

	// Try to load a file for reduced TD generation if it is present
	if _, err := os.Stat(reducedConfigFile); err == nil {
		byteValue, err := readByteValueFromJsonFile(reducedConfigFile)
		if err != nil {
			return err
		}
		err = json.Unmarshal(byteValue, &b.re)
		if err != nil {
			return err
		}
	}

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

	// Decide if a configuration file was applied
	if !configSet {
		// Apply own decisions on interaction affordance classification and ask the user if they are correct

		// translate the RPC functions into Interaction Affordances
		proto.Walk(definition,
			proto.WithRPC(b.HandleRPC))

		b.checkAndSaveInteractionAffordances()
	} else {
		byteValue, err := readByteValueFromJsonFile(classConfigFile)
		if err != nil {
			return err
		}
		err = json.Unmarshal(byteValue, &b.ac)
		if err != nil {
			return err
		}

		// Apply predefined configuration and classify affordances according to that
		proto.Walk(definition,
			proto.WithRPC(b.HandleRPCWithConfig))

		// Save affordances to the TD
		b.saveAfterConfigRPC()
	}

	// Generates a config file in the output directory to avoid terminal interaction for further generation
	b.generateConfigFileForAffordanceClassification(outputDir + "/classificationConfig.json")

	// serialize the TD to JSONLD
	tdBytes, _ := json.Marshal(b.td)
	f, err := os.Create(outputDir + "/td.jsonld")
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

// readByteValueFromJsonFile reads in a json file into byteValue
func readByteValueFromJsonFile(file string) ([]byte, error) {
	jsonFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer func(jsonFile *os.File) {
		jsonFile.Close()
	}(jsonFile)

	return ioutil.ReadAll(jsonFile)
}
