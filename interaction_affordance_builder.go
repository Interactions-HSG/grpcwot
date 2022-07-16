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
	combinedProp []combinedProperties
	prop         []affs
	action       []affs
	event        []affs
}

type combinedProperties struct {
	Name     string
	GetProp  affs
	SetProp  affs
	Category int // 0: read only; 1: write only; 2: readwrite
}

type affs struct {
	Name string
	Req  *wot.DataSchema
	Res  *wot.DataSchema
}

func newInteractionAffordanceBuilder(dsb *dataSchemaBuilder) *interactionAffordanceBuilder {
	return &interactionAffordanceBuilder{
		[]*proto.RPC{},
		map[string]affs{},
		dsb,
		affClasses{
			[]combinedProperties{},
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

// combine RPCs together with their DataSchemes. Checks if the proto file was valid with regards to the RPCs
func (b *interactionAffordanceBuilder) conformRPCs() error {
	b.affs = map[string]affs{}
	for _, v := range b.rpcs {
		if _, found := b.affs[v.Name]; found {
			return errors.New("Duplicate RPC name found in proto file for RPC Name " + v.Name)
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

// group together belonging properties if their request and response type as well as their name match
func (b *interactionAffordanceBuilder) groupProperties() {
	empty := affs{}
	for k, v := range b.affC.prop {
		if v == empty {
			continue
		}
		b.affC.prop[k] = empty
		b.checkPropertyCombination(v, "GET", "SET", true, empty)
		b.checkPropertyCombination(v, "SET", "GET", false, empty)
	}
}

// helper function for groupProperties
func (b *interactionAffordanceBuilder) checkPropertyCombination(p affs, s1, s2 string, isGet bool, empty affs) {
	pName := strings.ToUpper(p.Name)
	if strings.HasPrefix(pName, s1) {
		propName := p.Name[3:]
		var cand affs

		for k, v := range b.affC.prop {
			if v == empty {
				continue
			}
			if strings.HasPrefix(strings.ToUpper(v.Name), s2) && v.Name[3:] == propName {
				if (isGet && v.Req == p.Res) ||
					(!isGet && v.Res == p.Req) {
					cand = v
					b.affC.prop[k] = empty
					break
				} else {
					if isGet {
						b.affC.action = append(b.affC.action, v)
						b.affC.prop[k] = empty
					} else {
						b.affC.action = append(b.affC.action, p)
						return
					}
				}
			}
		}
		if isGet {
			b.affC.combinedProp = append(b.affC.combinedProp, combinedProperties{
				Name:     propName,
				GetProp:  p,
				SetProp:  cand,
				Category: getPropertyCategory(p.Name, cand.Name),
			})
		} else {
			b.affC.combinedProp = append(b.affC.combinedProp, combinedProperties{
				Name:     propName,
				GetProp:  cand,
				SetProp:  p,
				Category: getPropertyCategory(cand.Name, p.Name),
			})
		}
	}
}

// Determines the category for a properts (0: readonly, 1: writeonly, 2: readwrite)
func getPropertyCategory(get, set string) int {
	switch {
	case set == "":
		return 0
	case get == "":
		return 1
	default:
		return 2
	}
}

// checkCondition type that can build a filter for RPCs. The implementations of this type are scripted below
type checkCondition func(affs) bool

func defaultConfig(_ affs) bool {
	return true
}

func startsWithGetCaseInsensitive(a affs) bool {
	return strings.HasPrefix(strings.ToUpper(a.Name), "GET")
}

func startsWithSetCaseInsensitive(a affs) bool {
	return strings.HasPrefix(strings.ToUpper(a.Name), "SET")
}

func startsWithGet(a affs) bool {
	return strings.HasPrefix(a.Name, "Get")
}

func startsWithSet(a affs) bool {
	return strings.HasPrefix(a.Name, "Set")
}

func typeNotEmpty(t *wot.DataSchema) bool {
	return t.ObjectSchema != nil &&
		t.Properties != nil &&
		len(t.Properties) != 0
}

func hasReturnType(a affs) bool {
	return typeNotEmpty(a.Res)
}

func hasRequestType(a affs) bool {
	return typeNotEmpty(a.Req)
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

// Apply checkConditions and filter properties -> events -> actions
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

// HandleRPCWithConfig classifies RPC functions to interaction affordances based on a provided configuration
func (b *interactionAffordanceBuilder) categorizeRPCsWithConfig(ac map[string]affClassConfig) error {
	processed := make([]string, len(ac))
	i := 0
	for _, v := range b.affs {
		c, ok := ac[v.Name]
		if !ok {
			return errors.New("Could not find pre configured classification for RPC " + v.Name)
		}
		processed[i] = v.Name
		i++

		switch c.AffClass {
		case "property":
			b.affC.prop = append(b.affC.prop, v)
		case "action":
			b.affC.event = append(b.affC.event, v)
		case "event":
			b.affC.action = append(b.affC.action, v)
		default:
			return errors.New("Defined AffClass which is not possible " + c.AffClass)
		}
	}
	if i != len(ac)-1 {
		m := "Processed not all configs. Only the following RPCs were in the proto: "
		for _, e := range processed {
			m = m + e + ", "
		}
		return errors.New(m)
	}
	return nil
}

// generate Interaction Affordance based on checkConditions for classification
func generateInteractionAffordances(protoFile *proto.Proto, dsb *dataSchemaBuilder) (*interactionAffordanceBuilder, error) {
	b := newInteractionAffordanceBuilder(dsb)

	proto.Walk(protoFile,
		proto.WithRPC(b.HandleRPC))

	err := b.conformRPCs()
	if err != nil {
		return nil, err
	}

	b.categorizeRPCs()

	b.groupProperties()

	return b, nil
}

// generate Interaction Affordance when a configuration file is provided
func generateInteractionAffordancesWithConfig(protoFile *proto.Proto, dsb *dataSchemaBuilder, ac map[string]affClassConfig) (*interactionAffordanceBuilder, error) {
	b := newInteractionAffordanceBuilder(dsb)

	proto.Walk(protoFile,
		proto.WithRPC(b.HandleRPC))

	err := b.categorizeRPCsWithConfig(ac)
	if err != nil {
		return nil, err
	}
	return b, nil
}
