package gomib_test

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/golangsnmp/gomib"
	"github.com/golangsnmp/gomib/internal/testutil"
	"github.com/golangsnmp/gomib/mib"
)

// Exercise the public loading path; exhaustive ownership cases live in internal/model.
func loadImmutabilityMIB(t *testing.T) *mib.Mib {
	t.Helper()
	src := gomib.FS("immutability", fstest.MapFS{
		"IMMUTABILITY-MIB.mib": &fstest.MapFile{Data: []byte(`
IMMUTABILITY-MIB DEFINITIONS ::= BEGIN
IMPORTS
    MODULE-IDENTITY, OBJECT-TYPE, NOTIFICATION-TYPE, enterprises FROM SNMPv2-SMI
    MODULE-COMPLIANCE, OBJECT-GROUP, AGENT-CAPABILITIES FROM SNMPv2-CONF;

immutabilityMIB MODULE-IDENTITY
    LAST-UPDATED "202601010000Z"
    ORGANIZATION "Test"
    CONTACT-INFO "Test"
    DESCRIPTION "Model immutability tests."
    REVISION "202601010000Z"
    DESCRIPTION "Initial revision."
    ::= { enterprises 99999 }

bytesObject OBJECT-TYPE
    SYNTAX OCTET STRING (SIZE (1..8))
    MAX-ACCESS read-write
    STATUS current
    DESCRIPTION "Byte default."
    DEFVAL { 'AABB'H }
    ::= { immutabilityMIB 1 }

testNotification NOTIFICATION-TYPE
    OBJECTS { bytesObject }
    STATUS current
    DESCRIPTION "Notification."
    ::= { immutabilityMIB 2 }

testGroup OBJECT-GROUP
    OBJECTS { bytesObject }
    STATUS current
    DESCRIPTION "Objects."
    ::= { immutabilityMIB 3 }

testCompliance MODULE-COMPLIANCE
    STATUS current
    DESCRIPTION "Compliance."
    MODULE
        MANDATORY-GROUPS { testGroup }
        OBJECT bytesObject
            SYNTAX OCTET STRING (SIZE (2..3))
            MIN-ACCESS read-only
            DESCRIPTION "Restricted access."
    ::= { immutabilityMIB 4 }

testCapability AGENT-CAPABILITIES
    PRODUCT-RELEASE "Test"
    STATUS current
    DESCRIPTION "Capabilities."
    SUPPORTS IMMUTABILITY-MIB
        INCLUDES { testGroup }
        VARIATION bytesObject
            SYNTAX OCTET STRING (SIZE (2..3))
            ACCESS read-only
            DEFVAL { 'AABB'H }
            DESCRIPTION "Restricted access and byte default."
        VARIATION testNotification
            ACCESS not-implemented
            DESCRIPTION "Unsupported notification."
    ::= { immutabilityMIB 5 }
END
`)},
	})
	m, err := gomib.Load(context.Background(), gomib.WithSource(src), gomib.WithModules("IMMUTABILITY-MIB"))
	testutil.NoError(t, err, "Load")
	return m
}

func TestLoadedModelIsolation(t *testing.T) {
	for _, tc := range []struct {
		name  string
		check func(*testing.T, *mib.Mib)
	}{
		{"ObjectDefault", func(t *testing.T, m *mib.Mib) {
			d := requireObject(t, m, "bytesObject").DefaultValue()
			testutil.Isolated(t, []byte{0xAA, 0xBB}, func() []byte {
				v, ok := mib.DefValAs[[]byte](d)
				testutil.True(t, ok, "byte default")
				return v
			}, func(v []byte) { v[0] = 0 })
		}},
		{"VariationDefault", func(t *testing.T, m *mib.Mib) {
			c := m.Capability("testCapability")
			testutil.NotNil(t, c, "capability")
			d := c.Supports()[0].ObjectVariations[0].DefVal
			testutil.Isolated(t, []byte{0xAA, 0xBB}, func() []byte { return d.Value().([]byte) }, func(v []byte) { v[0] = 0 })
		}},
		{"Compliance", func(t *testing.T, m *mib.Mib) {
			c := m.Compliance("testCompliance")
			testutil.NotNil(t, c, "compliance")
			want := mib.AccessReadOnly
			testutil.Isolated(t, &want, func() *mib.Access { return c.Modules()[0].Objects[0].MinAccess }, func(v *mib.Access) { *v = mib.AccessReadWrite })
		}},
		{"CapabilityObject", func(t *testing.T, m *mib.Mib) {
			c := m.Capability("testCapability")
			testutil.NotNil(t, c, "capability")
			want := mib.AccessReadOnly
			testutil.Isolated(t, &want, func() *mib.Access { return c.Supports()[0].ObjectVariations[0].Access }, func(v *mib.Access) { *v = mib.AccessReadWrite })
		}},
		{"CapabilityNotification", func(t *testing.T, m *mib.Mib) {
			c := m.Capability("testCapability")
			testutil.NotNil(t, c, "capability")
			want := mib.AccessNotImplemented
			testutil.Isolated(t, &want, func() *mib.Access { return c.Supports()[0].NotificationVariations[0].Access }, func(v *mib.Access) { *v = mib.AccessReadOnly })
		}},
		{"ComplianceSyntax", func(t *testing.T, m *mib.Mib) {
			c := m.Compliance("testCompliance")
			testutil.NotNil(t, c, "compliance")
			testutil.SliceIsolated(t, func() []mib.Range { return c.Modules()[0].Objects[0].Syntax.Sizes })
		}},
		{"CapabilitySyntax", func(t *testing.T, m *mib.Mib) {
			c := m.Capability("testCapability")
			testutil.NotNil(t, c, "capability")
			testutil.SliceIsolated(t, func() []mib.Range { return c.Supports()[0].ObjectVariations[0].Syntax.Sizes })
		}},
	} {
		t.Run(tc.name, func(t *testing.T) { tc.check(t, loadImmutabilityMIB(t)) })
	}
}
