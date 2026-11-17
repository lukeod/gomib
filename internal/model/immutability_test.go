package model

import (
	"testing"

	"github.com/golangsnmp/gomib/internal/module"
	"github.com/golangsnmp/gomib/internal/testutil"
)

// Build directly so accessor ownership tests do not depend on parsing or resolution.
func buildCollectionMIB() *Mib {
	m := NewMib()
	mod := NewModule("TEST-MIB")
	m.addModule(mod)
	mod.rawImports = []module.Import{module.NewImport("OTHER-MIB", "value", Span{})}
	mod.revisions = []Revision{{Date: "202601010000Z", Description: "Initial revision"}}
	mod.oid = OID{1, 3}
	for i, kind := range []Kind{KindTable, KindRow, KindColumn, KindScalar} {
		n := &Node{arc: uint32(i + 1), parent: m.root, kind: kind}
		o := NewObject(kind.String())
		o.node, n.obj = n, o
		mod.addObject(o)
		m.addObject(o)
		mod.addNode(n)
		m.registerNode(o.name, n)
	}
	mod.addType(NewType("TestType"))
	mod.addNotification(NewNotification("testNotification"))
	mod.addGroup(NewGroup("testGroup"))
	mod.addCompliance(NewCompliance("testCompliance"))
	mod.addCapability(NewCapability("testCapability"))
	m.types, m.notifications, m.groups = mod.types, mod.notifications, mod.groups
	m.compliances, m.capabilities = mod.compliances, mod.capabilities
	return m
}

type modelCollections interface {
	Objects() []*Object
	Types() []*Type
	Notifications() []*Notification
	Groups() []*Group
	Compliances() []*Compliance
	Capabilities() []*Capability
	Tables() []*Object
	Rows() []*Object
	Columns() []*Object
	Scalars() []*Object
}

func TestModelCollectionIsolation(t *testing.T) {
	for _, scope := range []string{"Mib", "Module"} {
		for _, tc := range []struct {
			name  string
			check func(*testing.T, modelCollections)
		}{
			{"Objects", func(t *testing.T, c modelCollections) { testutil.SliceIsolated(t, c.Objects) }},
			{"Types", func(t *testing.T, c modelCollections) { testutil.SliceIsolated(t, c.Types) }},
			{"Notifications", func(t *testing.T, c modelCollections) { testutil.SliceIsolated(t, c.Notifications) }},
			{"Groups", func(t *testing.T, c modelCollections) { testutil.SliceIsolated(t, c.Groups) }},
			{"Compliances", func(t *testing.T, c modelCollections) { testutil.SliceIsolated(t, c.Compliances) }},
			{"Capabilities", func(t *testing.T, c modelCollections) { testutil.SliceIsolated(t, c.Capabilities) }},
			{"Tables", func(t *testing.T, c modelCollections) { testutil.SliceIsolated(t, c.Tables) }},
			{"Rows", func(t *testing.T, c modelCollections) { testutil.SliceIsolated(t, c.Rows) }},
			{"Columns", func(t *testing.T, c modelCollections) { testutil.SliceIsolated(t, c.Columns) }},
			{"Scalars", func(t *testing.T, c modelCollections) { testutil.SliceIsolated(t, c.Scalars) }},
		} {
			t.Run(scope+"/"+tc.name, func(t *testing.T) {
				m := buildCollectionMIB()
				var c modelCollections = m
				if scope == "Module" {
					c = m.Module("TEST-MIB")
				}
				tc.check(t, c)
			})
		}
	}
	for _, tc := range []struct {
		name  string
		check func(*testing.T, *Mib, *Module)
	}{
		{"Modules", func(t *testing.T, m *Mib, _ *Module) { testutil.SliceIsolated(t, m.Modules) }},
		{"ModuleNodes", func(t *testing.T, _ *Mib, mod *Module) { testutil.SliceIsolated(t, mod.Nodes) }},
		{"NodesByName", func(t *testing.T, m *Mib, _ *Module) {
			testutil.SliceIsolated(t, func() []*Node { return m.NodesByName(KindScalar.String()) })
		}},
		{"ModulesDefining", func(t *testing.T, m *Mib, _ *Module) {
			testutil.SliceIsolated(t, func() []*Module { return m.ModulesDefining(KindScalar.String()) })
		}},
		{"ModulesImporting", func(t *testing.T, m *Mib, _ *Module) {
			testutil.SliceIsolated(t, func() []*Module { return m.ModulesImporting("value") })
		}},
		{"ModuleOID", func(t *testing.T, _ *Mib, mod *Module) { testutil.SliceIsolated(t, mod.OID) }},
		{"Revisions", func(t *testing.T, _ *Mib, mod *Module) { testutil.SliceIsolated(t, mod.Revisions) }},
		{"Imports", func(t *testing.T, _ *Mib, mod *Module) { testutil.SliceIsolated(t, mod.Imports) }},
		{"ImportSymbols", func(t *testing.T, _ *Mib, mod *Module) {
			testutil.SliceIsolated(t, func() []ImportSymbol { return mod.Imports()[0].Symbols })
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := buildCollectionMIB()
			tc.check(t, m, m.Module("TEST-MIB"))
		})
	}
}

func TestObjectDataIsolation(t *testing.T) {
	for _, tc := range []struct {
		name  string
		check func(*testing.T, *Object)
	}{
		{"Sizes", func(t *testing.T, o *Object) { testutil.SliceIsolated(t, o.Sizes) }},
		{"EffectiveSizes", func(t *testing.T, o *Object) { testutil.SliceIsolated(t, o.EffectiveSizes) }},
		{"Ranges", func(t *testing.T, o *Object) { testutil.SliceIsolated(t, o.Ranges) }},
		{"EffectiveRanges", func(t *testing.T, o *Object) { testutil.SliceIsolated(t, o.EffectiveRanges) }},
		{"EffectiveEnums", func(t *testing.T, o *Object) { testutil.SliceIsolated(t, o.EffectiveEnums) }},
		{"EffectiveBits", func(t *testing.T, o *Object) { testutil.SliceIsolated(t, o.EffectiveBits) }},
		{"Index", func(t *testing.T, o *Object) { testutil.SliceIsolated(t, o.Index) }},
		{"EffectiveIndexes", func(t *testing.T, o *Object) { testutil.SliceIsolated(t, o.EffectiveIndexes) }},
		{"AugmentedBy", func(t *testing.T, o *Object) { testutil.SliceIsolated(t, o.AugmentedBy) }},
		{"ParsedDisplayHint", func(t *testing.T, o *Object) {
			testutil.SliceIsolated(t, func() []OctetSegment { return o.ParsedDisplayHint().OctetString().Segments })
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := &Object{
				entity:         entity{node: &Node{kind: KindRow}},
				sizes:          []Range{{Min: NewSignedRangeBound(1), Max: NewSignedRangeBound(8)}},
				declaredSizes:  []Range{{Min: NewSignedRangeBound(2), Max: NewSignedRangeBound(4)}},
				ranges:         []Range{{Min: NewSignedRangeBound(0), Max: NewSignedRangeBound(10)}},
				declaredRanges: []Range{{Min: NewSignedRangeBound(2), Max: NewSignedRangeBound(8)}},
				enums:          []NamedValue{{Label: "first", Value: 1}},
				bits:           []NamedValue{{Label: "first", Value: 0}},
				index:          []IndexEntry{{Object: NewObject("index")}},
				augmentedBy:    []*Object{NewObject("augmentEntry")},
				hint:           "1x:",
			}
			tc.check(t, o)
		})
	}
	t.Run("AugmentedIndexes", func(t *testing.T) {
		base := &Object{entity: entity{node: &Node{kind: KindRow}}, index: []IndexEntry{{Object: NewObject("index")}}}
		augment := &Object{entity: entity{node: &Node{kind: KindRow}}, augments: base}
		testutil.SliceIsolated(t, augment.EffectiveIndexes)
	})
}

func TestTypeDataIsolation(t *testing.T) {
	for _, inherited := range []bool{false, true} {
		scope := "Declared"
		if inherited {
			scope = "Inherited"
		}
		for _, tc := range []struct {
			name  string
			check func(*testing.T, *Type)
		}{
			{"Sizes", func(t *testing.T, ty *Type) { testutil.SliceIsolated(t, ty.Sizes) }},
			{"Ranges", func(t *testing.T, ty *Type) { testutil.SliceIsolated(t, ty.Ranges) }},
			{"Enums", func(t *testing.T, ty *Type) { testutil.SliceIsolated(t, ty.Enums) }},
			{"Bits", func(t *testing.T, ty *Type) { testutil.SliceIsolated(t, ty.Bits) }},
			{"EffectiveSizes", func(t *testing.T, ty *Type) { testutil.SliceIsolated(t, ty.EffectiveSizes) }},
			{"EffectiveRanges", func(t *testing.T, ty *Type) { testutil.SliceIsolated(t, ty.EffectiveRanges) }},
			{"EffectiveEnums", func(t *testing.T, ty *Type) { testutil.SliceIsolated(t, ty.EffectiveEnums) }},
			{"EffectiveBits", func(t *testing.T, ty *Type) { testutil.SliceIsolated(t, ty.EffectiveBits) }},
			{"ParsedDisplayHint", func(t *testing.T, ty *Type) {
				testutil.SliceIsolated(t, func() []OctetSegment { return ty.ParsedDisplayHint().OctetString().Segments })
			}},
		} {
			if inherited && (tc.name == "Sizes" || tc.name == "Ranges" || tc.name == "Enums" || tc.name == "Bits") {
				continue
			}
			t.Run(scope+"/"+tc.name, func(t *testing.T) {
				ty := &Type{
					sizes:  []Range{{Min: NewSignedRangeBound(1), Max: NewSignedRangeBound(8)}},
					ranges: []Range{{Min: NewSignedRangeBound(0), Max: NewSignedRangeBound(10)}},
					enums:  []NamedValue{{Label: "first", Value: 1}},
					bits:   []NamedValue{{Label: "first", Value: 0}},
					hint:   "1x:",
				}
				if inherited {
					ty = &Type{parent: ty}
				}
				tc.check(t, ty)
			})
		}
	}
}

func TestTreeDataIsolation(t *testing.T) {
	for _, tc := range []struct {
		name  string
		check func(*testing.T, *Node)
	}{
		{"Children", func(t *testing.T, root *Node) { testutil.SliceIsolated(t, root.Children) }},
		{"NodeOID", func(t *testing.T, root *Node) { testutil.SliceIsolated(t, root.Children()[0].OID) }},
		{"NotificationObjects", func(t *testing.T, _ *Node) {
			n := &Notification{objects: []*Object{NewObject("value")}}
			testutil.SliceIsolated(t, n.Objects)
		}},
		{"GroupMembers", func(t *testing.T, root *Node) {
			g := &Group{members: root.Children()}
			testutil.SliceIsolated(t, g.Members)
		}},
		{"RowColumns", func(t *testing.T, _ *Node) {
			row := &Node{kind: KindRow}
			col := &Node{kind: KindColumn, obj: NewObject("column"), parent: row}
			row.children = map[uint32]*Node{1: col}
			o := &Object{entity: entity{node: row}}
			testutil.SliceIsolated(t, o.Columns)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) { tc.check(t, buildTree()) })
	}
	// These accessors are promoted from entity into all five entity types.
	for _, field := range []string{"OID", "OidRefs"} {
		t.Run("Entity/"+field, func(t *testing.T) {
			e := &entity{node: buildTree().Children()[0], oidRefs: []OidRef{{Name: "parent"}}}
			if field == "OID" {
				testutil.SliceIsolated(t, e.OID)
			} else {
				testutil.SliceIsolated(t, e.OidRefs)
			}
		})
	}
	t.Run("TrapInfo", func(t *testing.T) {
		n := &Notification{trapInfo: &TrapInfo{Enterprise: "enterprises", TrapNumber: 1}}
		testutil.Isolated(t, &TrapInfo{Enterprise: "enterprises", TrapNumber: 1}, n.TrapInfo, func(info *TrapInfo) {
			info.Enterprise, info.TrapNumber = "changed", 99
		})
	})
}

func newSyntaxConstraints() *SyntaxConstraints {
	return &SyntaxConstraints{
		Type:              NewType("TestType"),
		Sizes:             []Range{{Min: NewSignedRangeBound(1), Max: NewSignedRangeBound(8)}},
		Ranges:            []Range{{Min: NewSignedRangeBound(0), Max: NewSignedRangeBound(10)}},
		DeclaredSizes:     []Range{{Min: NewSignedRangeBound(2), Max: NewSignedRangeBound(4)}},
		DeclaredRanges:    []Range{{Min: NewSignedRangeBound(2), Max: NewSignedRangeBound(8)}},
		SizesConstrained:  true,
		RangesConstrained: true,
		Enums:             []NamedValue{{Label: "first", Value: 1}},
		Bits:              []NamedValue{{Label: "first", Value: 0}},
	}
}

func newComplianceModule() ComplianceModule {
	access := AccessReadOnly
	return ComplianceModule{
		ModuleName:      "TEST-MIB",
		MandatoryGroups: []NameRef{{Name: "testGroup"}},
		Groups:          []ComplianceGroup{{Group: "testGroup", Description: "Optional group"}},
		Objects: []ComplianceObject{{
			Object: "value", Syntax: newSyntaxConstraints(), WriteSyntax: newSyntaxConstraints(), MinAccess: &access,
		}},
	}
}

func newCapabilitiesModule() CapabilitiesModule {
	objectAccess, notificationAccess := AccessReadOnly, AccessNotImplemented
	return CapabilitiesModule{
		ModuleName: "TEST-MIB",
		Includes:   []NameRef{{Name: "testGroup"}},
		ObjectVariations: []ObjectVariation{{
			Object: "value", Syntax: newSyntaxConstraints(), WriteSyntax: newSyntaxConstraints(), Access: &objectAccess,
			CreationRequires: []NameRef{{Name: "index"}},
			DefVal:           NewDefValBytes([]byte{0xAA, 0xBB}, "'AABB'H"),
		}},
		NotificationVariations: []NotificationVariation{{Notification: "testNotification", Access: &notificationAccess}},
	}
}

func TestComplianceMetadataIsolation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func([]ComplianceModule)
	}{
		{"Modules", func(v []ComplianceModule) { v[0] = ComplianceModule{} }},
		{"MandatoryGroups", func(v []ComplianceModule) { v[0].MandatoryGroups[0].Name = "changed" }},
		{"Groups", func(v []ComplianceModule) { v[0].Groups[0].Group = "changed" }},
		{"Objects", func(v []ComplianceModule) { v[0].Objects[0].Object = "changed" }},
		{"MinAccess", func(v []ComplianceModule) { *v[0].Objects[0].MinAccess = AccessReadWrite }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := &Compliance{modules: []ComplianceModule{newComplianceModule()}}
			testutil.Isolated(t, []ComplianceModule{newComplianceModule()}, c.Modules, tc.mutate)
		})
	}
}

func TestCapabilityMetadataIsolation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func([]CapabilitiesModule)
	}{
		{"Supports", func(v []CapabilitiesModule) { v[0] = CapabilitiesModule{} }},
		{"Includes", func(v []CapabilitiesModule) { v[0].Includes[0].Name = "changed" }},
		{"ObjectVariations", func(v []CapabilitiesModule) { v[0].ObjectVariations[0].Object = "changed" }},
		{"NotificationVariations", func(v []CapabilitiesModule) { v[0].NotificationVariations[0].Notification = "changed" }},
		{"CreationRequires", func(v []CapabilitiesModule) { v[0].ObjectVariations[0].CreationRequires[0].Name = "changed" }},
		{"ObjectAccess", func(v []CapabilitiesModule) { *v[0].ObjectVariations[0].Access = AccessReadWrite }},
		{"NotificationAccess", func(v []CapabilitiesModule) { *v[0].NotificationVariations[0].Access = AccessReadOnly }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := &Capability{supports: []CapabilitiesModule{newCapabilitiesModule()}}
			testutil.Isolated(t, []CapabilitiesModule{newCapabilitiesModule()}, c.Supports, tc.mutate)
		})
	}
}

func TestConformanceSyntaxIsolation(t *testing.T) {
	for _, source := range []struct {
		name    string
		newRead func() func() *SyntaxConstraints
	}{
		{"Compliance/Syntax", func() func() *SyntaxConstraints {
			c := &Compliance{modules: []ComplianceModule{newComplianceModule()}}
			return func() *SyntaxConstraints { return c.Modules()[0].Objects[0].Syntax }
		}},
		{"Compliance/WriteSyntax", func() func() *SyntaxConstraints {
			c := &Compliance{modules: []ComplianceModule{newComplianceModule()}}
			return func() *SyntaxConstraints { return c.Modules()[0].Objects[0].WriteSyntax }
		}},
		{"Capability/Syntax", func() func() *SyntaxConstraints {
			c := &Capability{supports: []CapabilitiesModule{newCapabilitiesModule()}}
			return func() *SyntaxConstraints { return c.Supports()[0].ObjectVariations[0].Syntax }
		}},
		{"Capability/WriteSyntax", func() func() *SyntaxConstraints {
			c := &Capability{supports: []CapabilitiesModule{newCapabilitiesModule()}}
			return func() *SyntaxConstraints { return c.Supports()[0].ObjectVariations[0].WriteSyntax }
		}},
	} {
		for _, tc := range []struct {
			name   string
			mutate func(*SyntaxConstraints)
		}{
			{"Struct", func(v *SyntaxConstraints) { v.Type = nil; v.SizesConstrained = false; v.RangesConstrained = false }},
			{"Sizes", func(v *SyntaxConstraints) { v.Sizes[0] = Range{} }},
			{"Ranges", func(v *SyntaxConstraints) { v.Ranges[0] = Range{} }},
			{"DeclaredSizes", func(v *SyntaxConstraints) { v.DeclaredSizes[0] = Range{} }},
			{"DeclaredRanges", func(v *SyntaxConstraints) { v.DeclaredRanges[0] = Range{} }},
			{"Enums", func(v *SyntaxConstraints) { v.Enums[0] = NamedValue{} }},
			{"Bits", func(v *SyntaxConstraints) { v.Bits[0] = NamedValue{} }},
		} {
			t.Run(source.name+"/"+tc.name, func(t *testing.T) {
				testutil.Isolated(t, newSyntaxConstraints(), source.newRead(), tc.mutate)
			})
		}
	}
}

func requireDefValIsolation[S ~[]E, E comparable](t *testing.T, newDefVal func() DefVal, want S) {
	t.Helper()
	for _, source := range []struct {
		name    string
		newRead func() func() DefVal
	}{
		{"Object", func() func() DefVal {
			d := newDefVal()
			o := &Object{defVal: &d}
			return o.DefaultValue
		}},
		{"Variation", func() func() DefVal {
			c := &Capability{supports: []CapabilitiesModule{{ObjectVariations: []ObjectVariation{{DefVal: newDefVal()}}}}}
			return func() DefVal { return c.Supports()[0].ObjectVariations[0].DefVal }
		}},
	} {
		for _, accessor := range []struct {
			name string
			read func(*testing.T, DefVal) S
		}{
			{"Value", func(t *testing.T, d DefVal) S { return d.Value().(S) }},
			{"DefValAs", func(t *testing.T, d DefVal) S {
				v, ok := DefValAs[S](d)
				testutil.True(t, ok, "DefValAs must succeed")
				return v
			}},
			{"DefValAsAny", func(t *testing.T, d DefVal) S {
				v, ok := DefValAs[any](d)
				testutil.True(t, ok, "DefValAs[any] must succeed")
				return v.(S)
			}},
		} {
			t.Run(source.name+"/"+accessor.name, func(t *testing.T) {
				read := source.newRead()
				d := read()
				raw, formatted, kind := d.Raw(), d.String(), d.Kind()
				testutil.Isolated(t, want, func() S { return accessor.read(t, d) }, func(v S) {
					var zero E
					v[0] = zero
				})
				// A saved DefVal and a new model read must both remain unchanged.
				for _, got := range []DefVal{d, read()} {
					testutil.SliceEqual(t, want, got.Value().(S), "default value")
					testutil.Equal(t, raw, got.Raw(), "raw default")
					testutil.Equal(t, formatted, got.String(), "formatted default")
					testutil.Equal(t, kind, got.Kind(), "default kind")
				}
			})
		}
	}
}

func TestDefaultValueIsolation(t *testing.T) {
	t.Run("Bytes", func(t *testing.T) {
		requireDefValIsolation(t, func() DefVal { return NewDefValBytes([]byte{0xAA, 0xBB}, "'AABB'H") }, []byte{0xAA, 0xBB})
	})
	t.Run("Bits", func(t *testing.T) {
		requireDefValIsolation(t, func() DefVal { return NewDefValBits([]string{"first", "second"}, "{ first, second }") }, []string{"first", "second"})
	})
	t.Run("OID", func(t *testing.T) {
		requireDefValIsolation(t, func() DefVal { return NewDefValOID(OID{1, 3, 6}, "parent") }, OID{1, 3, 6})
	})
}

func TestAbsentModelValues(t *testing.T) {
	t.Run("Object", func(t *testing.T) {
		o := NewObject("absent")
		d := o.DefaultValue()
		testutil.True(t, d.IsZero(), "absent default")
		testutil.Equal(t, DefValKindUnset, d.Kind(), "absent default kind")
		testutil.Nil(t, d.Value(), "absent default value")
		_, ok := DefValAs[[]byte](d)
		testutil.False(t, ok, "absent default must not yield bytes")
		testutil.Nil(t, o.Index(), "absent indexes")
		testutil.Nil(t, o.AugmentedBy(), "absent augmentations")
	})
	t.Run("Compliance", func(t *testing.T) {
		c := &Compliance{modules: []ComplianceModule{{Objects: []ComplianceObject{{Object: "value"}}}}}
		o := c.Modules()[0].Objects[0]
		testutil.Nil(t, o.MinAccess, "absent MIN-ACCESS")
		testutil.Nil(t, o.Syntax, "absent SYNTAX")
		testutil.Nil(t, o.WriteSyntax, "absent WRITE-SYNTAX")
	})
	t.Run("Capability", func(t *testing.T) {
		c := &Capability{supports: []CapabilitiesModule{{
			ObjectVariations:       []ObjectVariation{{Object: "value"}},
			NotificationVariations: []NotificationVariation{{Notification: "notification"}},
		}}}
		support := c.Supports()[0]
		o := support.ObjectVariations[0]
		testutil.Nil(t, o.Access, "absent object ACCESS")
		testutil.Nil(t, support.NotificationVariations[0].Access, "absent notification ACCESS")
		testutil.Nil(t, o.Syntax, "absent SYNTAX")
		testutil.Nil(t, o.WriteSyntax, "absent WRITE-SYNTAX")
		testutil.Nil(t, o.CreationRequires, "absent CREATION-REQUIRES")
	})
	t.Run("OID", func(t *testing.T) {
		testutil.Nil(t, NewMib().Root().OID(), "root OID")
		testutil.Nil(t, NewModule("absent").OID(), "absent module OID")
	})
}
