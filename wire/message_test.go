package wire

import (
	"bytes"
	"testing"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	e := &Encoder{}
	e.PutInt(-42)
	e.PutUint(0xdeadbeef)
	e.PutFixed(FixedFromFloat(1.5))
	e.PutString("hello")
	e.PutNullString()
	e.PutArray([]byte{1, 2, 3, 4, 5})
	e.PutObject(7)

	if len(e.buf)%4 != 0 {
		t.Fatalf("encoded length %d is not 32-bit aligned", len(e.buf))
	}

	d := &Decoder{buf: e.buf}
	if v, err := d.Int(); err != nil || v != -42 {
		t.Errorf("Int = %d, %v", v, err)
	}
	if v, err := d.Uint(); err != nil || v != 0xdeadbeef {
		t.Errorf("Uint = %#x, %v", v, err)
	}
	if v, err := d.Fixed(); err != nil || v.Float() != 1.5 {
		t.Errorf("Fixed = %v, %v", v, err)
	}
	if s, ok, err := d.String(); err != nil || !ok || s != "hello" {
		t.Errorf("String = %q, %v, %v", s, ok, err)
	}
	if s, ok, err := d.String(); err != nil || ok || s != "" {
		t.Errorf("null String = %q, %v, %v", s, ok, err)
	}
	if a, err := d.Array(); err != nil || !bytes.Equal(a, []byte{1, 2, 3, 4, 5}) {
		t.Errorf("Array = %v, %v", a, err)
	}
	if v, err := d.Object(); err != nil || v != 7 {
		t.Errorf("Object = %d, %v", v, err)
	}
	if d.remaining() != 0 {
		t.Errorf("%d bytes left over", d.remaining())
	}
}

func TestStringPadding(t *testing.T) {
	// Length includes the NUL; total is padded to 4 bytes.
	cases := []struct {
		s       string
		encoded int // bytes after the length word
	}{
		{"", 4},    // 1 byte (NUL) padded to 4
		{"abc", 4}, // 4 bytes exactly
		{"abcd", 8},
		{"abcdefg", 8},
	}
	for _, tc := range cases {
		e := &Encoder{}
		e.PutString(tc.s)
		if got := len(e.buf) - 4; got != tc.encoded {
			t.Errorf("PutString(%q) body = %d bytes, want %d", tc.s, got, tc.encoded)
		}
		d := &Decoder{buf: e.buf}
		s, ok, err := d.String()
		if err != nil || !ok || s != tc.s {
			t.Errorf("round trip %q = %q, %v, %v", tc.s, s, ok, err)
		}
		if d.remaining() != 0 {
			t.Errorf("PutString(%q): %d bytes left over", tc.s, d.remaining())
		}
	}
}

func TestDecodeTruncated(t *testing.T) {
	e := &Encoder{}
	e.PutString("hello world")
	d := &Decoder{buf: e.buf[:6]}
	if _, _, err := d.String(); err == nil {
		t.Error("decoding a truncated string did not fail")
	}
	d = &Decoder{buf: []byte{1, 2}}
	if _, err := d.Uint(); err == nil {
		t.Error("decoding a truncated uint did not fail")
	}
}

func TestFixedConversions(t *testing.T) {
	cases := []struct {
		f    float64
		want float64
	}{
		{0, 0}, {1, 1}, {-1, -1}, {1.5, 1.5}, {0.25, 0.25}, {-2.75, -2.75},
	}
	for _, tc := range cases {
		if got := FixedFromFloat(tc.f).Float(); got != tc.want {
			t.Errorf("FixedFromFloat(%v).Float() = %v", tc.f, got)
		}
	}
	if FixedFromInt(5).Int() != 5 {
		t.Errorf("FixedFromInt(5).Int() = %d", FixedFromInt(5).Int())
	}
	if FixedFromInt(-5).Int() != -5 {
		t.Errorf("FixedFromInt(-5).Int() = %d", FixedFromInt(-5).Int())
	}
}
