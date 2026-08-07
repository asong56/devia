package core

import "testing"

func TestConvertRadix_AutoDetectHex(t *testing.T) {
	r, err := ConvertRadix("0xFF", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Dec != "255" || r.Bin != "11111111" || r.Oct != "377" || r.Hex != "ff" {
		t.Errorf("ConvertRadix(0xFF) = %+v, want dec=255 bin=11111111 oct=377 hex=ff", r)
	}
}

func TestConvertRadix_AutoDetectOctal(t *testing.T) {
	r, err := ConvertRadix("0o10000", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Dec != "4096" {
		t.Errorf("ConvertRadix(0o10000).Dec = %s, want 4096", r.Dec)
	}
}

func TestConvertRadix_AutoDetectBinary(t *testing.T) {
	r, err := ConvertRadix("0b1000000000000", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Dec != "4096" {
		t.Errorf("ConvertRadix(0b1000000000000).Dec = %s, want 4096", r.Dec)
	}
}

func TestConvertRadix_PlainNumberDefaultsToDecimal(t *testing.T) {
	r, err := ConvertRadix("255", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Hex != "ff" {
		t.Errorf("ConvertRadix(255).Hex = %s, want ff", r.Hex)
	}
}

func TestConvertRadix_ExplicitFromBaseOverridesAutoDetect(t *testing.T) {
	// With fromBase explicitly set, no prefix is expected — "ff" is
	// read directly as base-16 digits, not auto-detected (there's no
	// 0x prefix here at all).
	r, err := ConvertRadix("ff", 16)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Dec != "255" {
		t.Errorf("ConvertRadix(ff, from=16).Dec = %s, want 255", r.Dec)
	}
}

func TestConvertRadix_LargeNumberDoesNotOverflow(t *testing.T) {
	// math/big is used specifically so this doesn't silently wrap the
	// way a fixed-width int64 conversion would. 2^100 comfortably
	// exceeds int64 (max ~9.2 * 10^18, i.e. < 2^63).
	r, err := ConvertRadix("1267650600228229401496703205376", 10) // 2^100
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantHex := "10000000000000000000000000" // 2^100 in hex = 1 followed by 25 zeros
	if r.Hex != wantHex {
		t.Errorf("ConvertRadix(2^100).Hex = %s, want %s", r.Hex, wantHex)
	}
}

func TestConvertRadix_InvalidNumber(t *testing.T) {
	_, err := ConvertRadix("not-a-number", 0)
	if err == nil {
		t.Fatal("expected an error for invalid input, got nil")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("CodeOf(err) = %d, want CodeInput (%d)", CodeOf(err), CodeInput)
	}
}

func TestConvertRadix_DigitOutOfRangeForBase(t *testing.T) {
	// "9" is not a valid binary digit — this must be rejected, not
	// silently truncated or misparsed.
	_, err := ConvertRadix("9", 2)
	if err == nil {
		t.Fatal("expected an error for a digit that's invalid in the given base, got nil")
	}
}

func TestConvertRadix_ZeroAndNegative(t *testing.T) {
	r, err := ConvertRadix("0", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Dec != "0" || r.Hex != "0" {
		t.Errorf("ConvertRadix(0) = %+v, want all-zero representations", r)
	}

	neg, err := ConvertRadix("-255", 0)
	if err != nil {
		t.Fatalf("unexpected error converting a negative number: %v", err)
	}
	if neg.Dec != "-255" {
		t.Errorf("ConvertRadix(-255).Dec = %s, want -255", neg.Dec)
	}
}
