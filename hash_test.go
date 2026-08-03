package core

import (
	"encoding/base64"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

// These are standard, widely published test vectors (FIPS 180-4 / RFC
// 1321), not values computed by this test — used to catch a wrong
// algorithm being wired up rather than to re-derive cryptography.
func TestHashTextKnownVectors(t *testing.T) {
	cases := []struct {
		algo string
		text string
		want string
	}{
		{"md5", "", "d41d8cd98f00b204e9800998ecf8427e"},
		{"md5", "abc", "900150983cd24fb0d6963f7d28e17f72"},
		{"sha1", "", "da39a3ee5e6b4b0d3255bfef95601890afd80709"},
		{"sha1", "abc", "a9993e364706816aba3e25717850c26c9cd0d89d"},
		{"sha256", "", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"sha256", "abc", "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"},
	}
	for _, c := range cases {
		got, err := HashText(c.algo, c.text, "", false)
		if err != nil {
			t.Fatalf("HashText(%s, %q): %v", c.algo, c.text, err)
		}
		if got != c.want {
			t.Errorf("HashText(%s, %q) = %s, want %s", c.algo, c.text, got, c.want)
		}
	}
}

func TestHashTextDefaultsToSHA256(t *testing.T) {
	withAlgo, err := HashText("sha256", "hello", "", false)
	if err != nil {
		t.Fatal(err)
	}
	withEmpty, err := HashText("", "hello", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if withAlgo != withEmpty {
		t.Errorf("empty algo should default to sha256: got %s vs %s", withEmpty, withAlgo)
	}
}

func TestHashTextUnsupportedAlgo(t *testing.T) {
	_, err := HashText("crc32", "x", "", false)
	if err == nil {
		t.Fatal("expected an error for an unsupported algorithm")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("unsupported algorithm should be a CodeInput error, got code %d", CodeOf(err))
	}
}

func TestHashTextBase64MatchesHex(t *testing.T) {
	hexSum, err := HashText("sha256", "round trip me", "", false)
	if err != nil {
		t.Fatal(err)
	}
	b64Sum, err := HashText("sha256", "round trip me", "", true)
	if err != nil {
		t.Fatal(err)
	}

	rawFromHex, err := hex.DecodeString(hexSum)
	if err != nil {
		t.Fatalf("hex digest wasn't valid hex: %v", err)
	}
	rawFromB64, err := base64.StdEncoding.DecodeString(b64Sum)
	if err != nil {
		t.Fatalf("base64 digest wasn't valid base64: %v", err)
	}
	if string(rawFromHex) != string(rawFromB64) {
		t.Error("hex and base64 encodings of the same hash should decode to the same bytes")
	}
}

func TestHashTextHMACDiffersFromPlainHash(t *testing.T) {
	plain, err := HashText("sha256", "message", "", false)
	if err != nil {
		t.Fatal(err)
	}
	hmac1, err := HashText("sha256", "message", "key1", false)
	if err != nil {
		t.Fatal(err)
	}
	hmac2, err := HashText("sha256", "message", "key2", false)
	if err != nil {
		t.Fatal(err)
	}

	if plain == hmac1 {
		t.Error("HMAC output should differ from the plain hash")
	}
	if hmac1 == hmac2 {
		t.Error("HMAC output should depend on the key")
	}

	// Same algo+text+key should be fully deterministic.
	hmac1Again, _ := HashText("sha256", "message", "key1", false)
	if hmac1 != hmac1Again {
		t.Error("HMAC should be deterministic for the same inputs")
	}
}

func TestHashBytesMatchesHashText(t *testing.T) {
	text := "bytes and text should agree"
	viaText, err := HashText("sha512", text, "", false)
	if err != nil {
		t.Fatal(err)
	}
	viaBytes, err := HashBytes("sha512", []byte(text), false)
	if err != nil {
		t.Fatal(err)
	}
	if viaText != viaBytes {
		t.Errorf("HashText and HashBytes disagree: %s vs %s", viaText, viaBytes)
	}
}

func TestHashFileMatchesHashBytes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.bin")
	content := []byte("the quick brown fox jumps over the lazy dog")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	fromFile, err := HashFile("sha256", path, false)
	if err != nil {
		t.Fatal(err)
	}
	fromBytes, err := HashBytes("sha256", content, false)
	if err != nil {
		t.Fatal(err)
	}
	if fromFile != fromBytes {
		t.Errorf("HashFile and HashBytes disagree: %s vs %s", fromFile, fromBytes)
	}
}

func TestHashFileMissing(t *testing.T) {
	_, err := HashFile("sha256", filepath.Join(t.TempDir(), "does-not-exist"), false)
	if err == nil {
		t.Fatal("expected an error for a missing file")
	}
	if CodeOf(err) != CodeNotFound {
		t.Errorf("missing file should be a CodeNotFound error, got code %d", CodeOf(err))
	}
}
