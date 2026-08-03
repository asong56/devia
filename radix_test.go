package core

import "testing"

func TestConvertRadixDecimalInput(t *testing.T) {
	got, err := ConvertRadix("255", 0)
	if err != nil {
		t.Fatal(err)
	}
	want := RadixResult{Bin: "11111111", Oct: "377", Dec: "255", Hex: "ff"}
	if *got != want {
		t.Errorf("ConvertRadix(255) = %+v, want %+v", *got, want)
	}
}

func TestConvertRadixAutoDetectsPrefixes(t *testing.T) {
	cases := []struct {
		input string
		want  RadixResult
	}{
		{"0xFF", RadixResult{Bin: "11111111", Oct: "377", Dec: "255", Hex: "ff"}},
		{"0o377", RadixResult{Bin: "11111111", Oct: "377", Dec: "255", Hex: "ff"}},
		{"0b11111111", RadixResult{Bin: "11111111", Oct: "377", Dec: "255", Hex: "ff"}},
	}
	for _, c := range cases {
		got, err := ConvertRadix(c.input, 0)
		if err != nil {
			t.Fatalf("ConvertRadix(%s): %v", c.input, err)
		}
		if *got != c.want {
			t.Errorf("ConvertRadix(%s) = %+v, want %+v", c.input, *got, c.want)
		}
	}
}

func TestConvertRadixExplicitFromBase(t *testing.T) {
	// fromBase=16 with no 0x prefix.
	got, err := ConvertRadix("ff", 16)
	if err != nil {
		t.Fatal(err)
	}
	if got.Dec != "255" {
		t.Errorf("ConvertRadix(ff, base 16) Dec = %s, want 255", got.Dec)
	}
}

func TestConvertRadixLargeNumberDoesNotOverflow(t *testing.T) {
	// Larger than a 64-bit integer can hold, to exercise math/big.
	got, err := ConvertRadix("123456789012345678901234567890", 10)
	if err != nil {
		t.Fatal(err)
	}
	if got.Dec != "123456789012345678901234567890" {
		t.Errorf("large decimal round trip failed: got %s", got.Dec)
	}
}

func TestConvertRadixInvalidInput(t *testing.T) {
	_, err := ConvertRadix("not-a-number", 0)
	if err == nil {
		t.Fatal("expected an error for invalid input")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("invalid input should be a CodeInput error, got code %d", CodeOf(err))
	}
}

func TestConvertRadixZero(t *testing.T) {
	got, err := ConvertRadix("0", 0)
	if err != nil {
		t.Fatal(err)
	}
	want := RadixResult{Bin: "0", Oct: "0", Dec: "0", Hex: "0"}
	if *got != want {
		t.Errorf("ConvertRadix(0) = %+v, want %+v", *got, want)
	}
}
