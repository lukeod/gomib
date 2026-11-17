package model

import (
	"testing"

	"github.com/golangsnmp/gomib/internal/testutil"
)

func TestDefValString(t *testing.T) {
	tests := []struct {
		name string
		dv   DefVal
		want string
	}{
		{"int positive", NewDefValInt(42, "42"), "42"},
		{"int negative", NewDefValInt(-1, "-1"), "-1"},
		{"int zero", NewDefValInt(0, "0"), "0"},
		{"uint", NewDefValUint(100, "100"), "100"},
		{"uint zero", NewDefValUint(0, "0"), "0"},
		{"string", NewDefValString("hello", `"hello"`), `"hello"`},
		{"string empty", NewDefValString("", `""`), `""`},
		{"string with quotes", NewDefValString(`say "hi"`, `"say \"hi\""`), `"say \"hi\""`},
		{"enum label", NewDefValEnum("active", "active"), "active"},
		{"bits multiple", NewDefValBits([]string{"read", "write"}, "{ read, write }"), "{ read, write }"},
		{"bits empty", NewDefValBits([]string{}, "{ }"), "{ }"},
		{"bits single", NewDefValBits([]string{"read"}, "{ read }"), "{ read }"},
		{"oid", NewDefValOID(OID{0, 0}, "zeroDotZero"), "zeroDotZero"},
		{"bytes empty", NewDefValBytes([]byte{}, "''H"), "0x"},
		{"bytes 1 byte", NewDefValBytes([]byte{0xFF}, "'FF'H"), "0xFF"},
		{"bytes 4 bytes", NewDefValBytes([]byte{0xDE, 0xAD, 0xBE, 0xEF}, "'DEADBEEF'H"), "0xDEADBEEF"},
		{"bytes 8 bytes", NewDefValBytes([]byte{0, 0, 0, 0, 0, 0, 0, 1}, "'0000000000000001'H"), "0x0000000000000001"},
		{"bytes >8 bytes", NewDefValBytes([]byte{0xAB, 0xCD, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, "x"), "0xABCD00000000000001"},
		{"bytes all zero >8", NewDefValBytes(make([]byte, 16), "x"), "0x00000000000000000000000000000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.dv.String()
			testutil.Equal(t, tt.want, got, "DefVal.String()")
		})
	}
}

func TestDefValKindString(t *testing.T) {
	tests := []struct {
		kind DefValKind
		want string
	}{
		{DefValKindUnset, "unset"},
		{DefValKindInt, "int"},
		{DefValKindUint, "uint"},
		{DefValKindString, "string"},
		{DefValKindBytes, "bytes"},
		{DefValKindEnum, "enum"},
		{DefValKindBits, "bits"},
		{DefValKindOID, "oid"},
		{DefValKind(99), "DefValKind(99)"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.kind.String()
			testutil.Equal(t, tt.want, got, "DefValKind().String()")
		})
	}
}

func TestDefValIsZero(t *testing.T) {
	var zero DefVal
	testutil.True(t, zero.IsZero(), "zero DefVal should report IsZero() true")
	testutil.Equal(t, DefValKindUnset, zero.Kind(), "zero DefVal Kind()")

	nonZero := NewDefValInt(0, "0")
	testutil.False(t, nonZero.IsZero(), "NewDefValInt(0) should not be IsZero (value is set, just happens to be 0)")
}

func TestDefValKind(t *testing.T) {
	tests := []struct {
		name string
		dv   DefVal
		want DefValKind
	}{
		{"int", NewDefValInt(1, "1"), DefValKindInt},
		{"uint", NewDefValUint(1, "1"), DefValKindUint},
		{"string", NewDefValString("x", "x"), DefValKindString},
		{"bytes", NewDefValBytes([]byte{1}, "x"), DefValKindBytes},
		{"enum", NewDefValEnum("x", "x"), DefValKindEnum},
		{"bits", NewDefValBits([]string{"x"}, "x"), DefValKindBits},
		{"oid", NewDefValOID(OID{1}, "1"), DefValKindOID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.Equal(t, tt.want, tt.dv.Kind(), "Kind()")
		})
	}
}

func TestDefValValue(t *testing.T) {
	dv := NewDefValInt(42, "42")
	v := dv.Value()
	testutil.Equal(t, 42, v.(int64), "Value()")
}

func TestDefValRaw(t *testing.T) {
	dv := NewDefValInt(42, "42")
	testutil.Equal(t, "42", dv.Raw(), "Raw()")
}

func TestDefValAs(t *testing.T) {
	t.Run("int64 match", func(t *testing.T) {
		dv := NewDefValInt(42, "42")
		v, ok := DefValAs[int64](dv)
		testutil.True(t, ok, "DefValAs[int64] should succeed")
		testutil.Equal(t, 42, v, "got")
	})

	t.Run("uint64 match", func(t *testing.T) {
		dv := NewDefValUint(100, "100")
		v, ok := DefValAs[uint64](dv)
		testutil.True(t, ok, "DefValAs[uint64] should succeed")
		testutil.Equal(t, 100, v, "got")
	})

	t.Run("string match", func(t *testing.T) {
		dv := NewDefValString("hello", `"hello"`)
		v, ok := DefValAs[string](dv)
		testutil.True(t, ok, "DefValAs[string] should succeed")
		testutil.Equal(t, "hello", v, "got")
	})

	t.Run("bytes match", func(t *testing.T) {
		dv := NewDefValBytes([]byte{0xAB}, "x")
		v, ok := DefValAs[[]byte](dv)
		testutil.True(t, ok, "DefValAs[[]byte] should succeed")
		testutil.Len(t, v, 1, "DefValAs[[]byte] length")
		testutil.Equal(t, byte(0xAB), v[0], "DefValAs[[]byte] value")
	})

	t.Run("type mismatch", func(t *testing.T) {
		dv := NewDefValInt(42, "42")
		_, ok := DefValAs[string](dv)
		testutil.False(t, ok, "DefValAs[string] on int DefVal should return false")
	})

	t.Run("bits match", func(t *testing.T) {
		dv := NewDefValBits([]string{"a", "b"}, "x")
		v, ok := DefValAs[[]string](dv)
		testutil.True(t, ok, "DefValAs[[]string] should succeed")
		testutil.Len(t, v, 2, "got")
	})

	t.Run("oid match", func(t *testing.T) {
		dv := NewDefValOID(OID{1, 3}, "1.3")
		v, ok := DefValAs[OID](dv)
		testutil.True(t, ok, "DefValAs[OID] should succeed")
		testutil.Equal(t, "1.3", v.String(), "got")
	})
}

func TestDefValStringZeroValue(t *testing.T) {
	var zero DefVal
	testutil.True(t, zero.IsZero(), "zero DefVal should report IsZero() true")
	got := zero.String()
	testutil.Equal(t, "", got, "zero DefVal.String()")
}

func TestRangeString(t *testing.T) {
	tests := []struct {
		name string
		r    Range
		want string
	}{
		{"single value", Range{Min: NewSignedRangeBound(5), Max: NewSignedRangeBound(5)}, "5"},
		{"range", Range{Min: NewSignedRangeBound(0), Max: NewSignedRangeBound(255)}, "0..255"},
		{"negative", Range{Min: NewSignedRangeBound(-1), Max: NewSignedRangeBound(100)}, "-1..100"},
		{"zero range", Range{Min: NewSignedRangeBound(0), Max: NewSignedRangeBound(0)}, "0"},
		{"raw single value", Range{Min: NewRawRangeBound("'0G'H"), Max: NewRawRangeBound("'0G'H")}, "'0G'H"},
		{"raw range", Range{Min: NewRawRangeBound("'0G'H"), Max: NewRawRangeBound("'1G'H")}, "'0G'H..'1G'H"},
		{"raw lower endpoint", Range{Min: NewRawRangeBound("'0G'H"), Max: NewSignedRangeBound(10)}, "'0G'H..10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.r.String()
			testutil.Equal(t, tt.want, got, "Range.String()")
		})
	}
}

func TestBytesToHex(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{"single byte", []byte{0xAB}, "AB"},
		{"multiple", []byte{0xDE, 0xAD}, "DEAD"},
		{"zeros", []byte{0x00, 0x00}, "0000"},
		{"all FF", []byte{0xFF, 0xFF}, "FFFF"},
		{"empty", []byte{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bytesToHex(tt.input)
			testutil.Equal(t, tt.want, got, "bytesToHex()")
		})
	}
}

func requireEmptyDefault[S ~[]E, E any](t *testing.T, d DefVal, kind DefValKind) {
	t.Helper()
	testutil.False(t, d.IsZero(), "empty default is set")
	testutil.Equal(t, kind, d.Kind(), "empty default kind")
	value, ok := d.Value().(S)
	testutil.True(t, ok, "Value retains slice type")
	testutil.Len(t, value, 0, "empty Value")
	value, ok = DefValAs[S](d)
	testutil.True(t, ok, "DefValAs retains slice type")
	testutil.Len(t, value, 0, "empty DefValAs")
}

func TestDefValEmptySlices(t *testing.T) {
	t.Run("Bytes", func(t *testing.T) {
		requireEmptyDefault[[]byte](t, NewDefValBytes([]byte{}, "''H"), DefValKindBytes)
	})
	t.Run("Bits", func(t *testing.T) {
		requireEmptyDefault[[]string](t, NewDefValBits([]string{}, "{ }"), DefValKindBits)
	})
}
