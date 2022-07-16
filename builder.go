package grpcwot

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/emicklei/proto"
	"github.com/linksmart/thing-directory/wot"
)

type builder struct {
	td          wot.ThingDescription
	dsb         *dataSchemaBuilder
	iab         *interactionAffordanceBuilder
	ip          string
	port        int
	ac          map[string]affClassConfig
	handleError error
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

// saveToAffClass is a helper function to save affordances in the affordance classification, so they could be
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
	switch p.Category {
	case 0:
		b.saveToAffClass(p.GetProp.Name, p.Name, "property")
		affordance.DataSchema = *p.GetProp.Res
		ops = []string{"readproperty"}
	case 1:
		b.saveToAffClass(p.SetProp.Name, p.Name, "property")
		affordance.DataSchema = *p.SetProp.Req
		ops = []string{"writeproperty"}
	case 2:
		b.saveToAffClass(p.GetProp.Name, p.Name, "property")
		b.saveToAffClass(p.SetProp.Name, p.Name, "property")
		affordance.DataSchema = *p.GetProp.Res
		ops = []string{"readproperty", "writeproperty"}
	default:
		return
	}

	affordance.Forms = b.getForms(p.Name, ops)
	b.td.Properties[p.Name] = affordance
}

// saveAction converts and saves a RPC function to an Action Affordance in the TD
func (b *builder) saveAction(r affs) {
	affordance := wot.ActionAffordance{}
	affordance.Input = *r.Req
	affordance.Output = *r.Res
	affordance.Forms = b.getForms(r.Name, []string{})
	b.td.Actions[r.Name] = affordance

	b.saveToAffClass(r.Name, r.Name, "action")
}

// saveEvent converts and saves a RPC function to an Event Affordance in the TD
func (b *builder) saveEvent(r affs) {
	affordance := wot.EventAffordance{}
	affordance.Data = *r.Res
	affordance.Forms = b.getForms(r.Name, []string{})
	b.td.Events[r.Name] = affordance

	b.saveToAffClass(r.Name, r.Name, "event")
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
func (b *builder) categorizeAffordances() {
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
		t := readInput(reader, "Property", v.Name, []string{v.GetProp.Name, v.SetProp.Name})
		switch t {
		case "":
			fallthrough
		case "p":
			b.saveProperty(v)
		case "a":
			switch v.Category {
			case 2:
				fmt.Println("Should only the setter become an action (set) or both (both)?")
				fmt.Print("->")
				t, _ := reader.ReadString('\n')
				t = strings.TrimSpace(t)
				if t == "set" {
					b.saveAction(v.SetProp)
					v.SetProp = affs{}
					v.Category = 0
					b.saveProperty(v)
				} else if t == "both" {
					b.saveAction(v.GetProp)
					b.saveAction(v.SetProp)
				}
			case 1:
				b.saveAction(v.SetProp)
			case 0:
				b.saveAction(v.GetProp)
			}
		case "e":
			if v.Category == 2 {
				v.Category = 1
				b.saveProperty(v)
			}
			b.saveEvent(v.GetProp)
		}
	}
	if len(b.iab.affC.action) != 0 {
		fmt.Println("The following were considered as actions: ")
	}
	for _, v := range b.iab.affC.action {
		t := readInput(reader, "Action", v.Name, []string{v.Name})
		switch t {
		case "":
			fallthrough
		case "a":
			b.saveAction(v)
		case "p":
			if v.Req.Type == "" {
				b.saveProperty(combinedProperties{
					Name:     v.Name,
					GetProp:  v,
					Category: 0,
				})
			} else {
				b.saveProperty(combinedProperties{
					Name:     v.Name,
					SetProp:  v,
					Category: 1,
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
		t := readInput(reader, "Event", v.Name, []string{v.Name})
		switch t {
		case "":
			fallthrough
		case "e":
			b.saveEvent(v)
		case "p":
			b.saveProperty(combinedProperties{
				Name:     v.Name,
				GetProp:  v,
				Category: 0,
			})
		case "a":
			b.saveAction(v)
		}
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

// saveAfterConfigRPC Saves the affordances to the TD with the classification derived from the vonfig
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
func GenerateTDfromProtoBuf(protoFile, outputDir, classConfigFile, ip string, port int) error { // parse the protoFile with the emicklei/proto
	configSet := true
	// Check if config File is present
	if _, err := os.Stat(classConfigFile); errors.Is(err, os.ErrNotExist) {
		configSet = false
	}

	reader, _ := os.Open(protoFile)
	defer reader.Close()

	b, err := fillBuilder(reader, ip, port, configSet, false, classConfigFile)
	if err != nil {
		return err
	}

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

type serverAffordances struct {
	Props   []serverProperty
	Actions []serverAffordance
	Events  []serverAffordance
}

type serverProperty struct {
	Name     string
	GetProp  serverAffordance
	SetProp  serverAffordance
	Category int
}

type serverAffordance struct {
	Name string
	Req  serverDataSchema
	Res  serverDataSchema
}

type serverDataSchema struct {
	Type       string
	Properties []serverProp
}

type serverProp struct {
	Key   string
	Value serverDataSchema
}

func GetProtoBufInformation(protofile io.Reader) ([]byte, error) {
	b, err := fillBuilder(protofile, "", 0, false, true, "")
	if err != nil {
		return []byte{}, err
	}
	res2 := serverAffordances{
		Props:   []serverProperty{},
		Actions: []serverAffordance{},
		Events:  []serverAffordance{},
	}
	for _, elem := range b.iab.affC.combinedProp {
		res2.Props = append(res2.Props, serverProperty{
			Name: elem.Name,
			GetProp: serverAffordance{
				Name: elem.GetProp.Name,
				Req:  createServerDataSchema(elem.GetProp.Req),
				Res:  createServerDataSchema(elem.GetProp.Res),
			},
			SetProp: serverAffordance{
				Name: elem.SetProp.Name,
				Req:  createServerDataSchema(elem.SetProp.Req),
				Res:  createServerDataSchema(elem.SetProp.Res),
			},
			Category: elem.Category,
		})
	}
	for _, elem := range b.iab.affC.action {
		res2.Actions = append(res2.Actions, serverAffordance{
			Name: elem.Name,
			Req:  createServerDataSchema(elem.Req),
			Res:  createServerDataSchema(elem.Res),
		})
	}
	for _, elem := range b.iab.affC.event {
		res2.Events = append(res2.Events, serverAffordance{
			Name: elem.Name,
			Req:  createServerDataSchema(elem.Req),
			Res:  createServerDataSchema(elem.Res),
		})
	}

	result, err := json.Marshal(res2)
	if err != nil {
		return []byte{}, err
	}

	return result, nil
}

func createServerDataSchema(ds *wot.DataSchema) serverDataSchema {
	if ds == nil {
		return serverDataSchema{}
	}
	if ds.ObjectSchema == nil || len(ds.Properties) == 0 {
		return serverDataSchema{
			Type: ds.DataType,
		}
	} else {
		props := make([]serverProp, len(ds.Properties))
		i := 0
		for k, v := range ds.Properties {
			props[i] = serverProp{
				k,
				createServerDataSchema(&v),
			}
			i++
		}
		return serverDataSchema{
			Type:       ds.DataType,
			Properties: props,
		}
	}

}

func fillBuilder(reader io.Reader, ip string, port int, configSet, isServer bool, classConfigFile string) (*builder, error) {
	parser := proto.NewParser(reader)
	definition, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	// Read the Messages and produce DataSchemes
	dsb, err := generateDataSchemas(definition)
	if err != nil {
		return nil, err
	}

	// initialize the TD builder with an empty TD and DataSchema
	b := newBuilder(ip, port, dsb)

	// translate the RPC functions into Interaction Affordances
	proto.Walk(definition,
		proto.WithService(b.HandleService))

	if configSet {
		byteValue, err := readByteValueFromJsonFile(classConfigFile)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(byteValue, &b.ac)
		if err != nil {
			return nil, err
		}

		// Apply predefined configuration and classify affordances according to that
		b.iab, err = generateInteractionAffordancesWithConfig(definition, dsb, b.ac)

		// Save affordances to the TD
		b.saveAfterConfigRPC()
	} else {
		b.iab, err = generateInteractionAffordances(definition, dsb)
		if !isServer {
			b.categorizeAffordances()
		}
	}

	if b.iab == nil {
		return b, err
	}
	return b, nil
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
