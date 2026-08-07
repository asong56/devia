package core

import (
	"os"
	"path/filepath"
	"testing"
)

// Known-answer vectors — the classic empty-string and "abc" test
// vectors published for each algorithm, so these tests catch a wrong
// algorithm being wired up, not just "it returned something".
func TestHashText_KnownVectors(t *testing.T) {
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
		{"sha512", "abc", "ddaf35a193617abacc417349ae20413112e6fa4e89a97ea20a9eeee64b55d39a2192992a274fc1a836ba3c23a3feebbd454d4423643ce80e2a9ac94fa54ca49f"},
		{"", "abc", "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"}, // "" defaults to sha256
	}
	for _, c := range cases {
		got, err := HashText(c.algo, c.text, "", false)
		if err != nil {
			t.Errorf("HashText(%q, %q): unexpected error: %v", c.algo, c.text, err)
			continue
		}
		if got != c.want {
			t.Errorf("HashText(%q, %q) = %s, want %s", c.algo, c.text, got, c.want)
		}
	}
}

func TestHashText_UnsupportedAlgo(t *testing.T) {
	_, err := HashText("crc32", "x", "", false)
	if err == nil {
		t.Fatal("expected an error for an unsupported algorithm, got nil")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("CodeOf(err) = %d, want CodeInput (%d) — bad algo name is a user input error, not internal", CodeOf(err), CodeInput)
	}
}

func TestHashText_Base64Output(t *testing.T) {
	// md5("abc") hex = 900150983cd24fb0d6963f7d28e17f72
	// same bytes, base64-encoded:
	got, err := HashText("md5", "abc", "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "kAFQmDzST7DWlj99KOF/cg=="
	if got != want {
		t.Errorf("HashText(base64=true) = %s, want %s", got, want)
	}
}

func TestHashText_HMAC(t *testing.T) {
	// RFC 2202 HMAC-MD5 test case 1: key = 0x0b*16, data = "Hi There"
	key := string([]byte{0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b})
	got, err := HashText("md5", "Hi There", key, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "9294727a3638bb1c13f48ef8158bfc9d"
	if got != want {
		t.Errorf("HMAC-MD5 = %s, want %s (RFC 2202 test case 1)", got, want)
	}
}

func TestHashText_HMACDiffersFromPlainHash(t *testing.T) {
	plain, _ := HashText("sha256", "message", "", false)
	hmacd, _ := HashText("sha256", "message", "secret", false)
	if plain == hmacd {
		t.Error("HMAC output must not equal the plain hash of the same text")
	}
}

func TestHashBytes_MatchesHashText(t *testing.T) {
	// HashBytes and HashText should agree on the same content — one
	// reads a string, the other a []byte, but the hasher underneath
	// is identical.
	text := "the quick brown fox"
	byText, err := HashText("sha256", text, "", false)
	if err != nil {
		t.Fatalf("HashText error: %v", err)
	}
	byBytes, err := HashBytes("sha256", []byte(text), false)
	if err != nil {
		t.Fatalf("HashBytes error: %v", err)
	}
	if byText != byBytes {
		t.Errorf("HashText and HashBytes disagree: %s vs %s", byText, byBytes)
	}
}

func TestHashFile_MatchesHashText(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.txt")
	content := "streamed file content for hashing\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	want, _ := HashText("sha256", content, "", false)
	got, err := HashFile("sha256", path, false)
	if err != nil {
		t.Fatalf("HashFile error: %v", err)
	}
	if got != want {
		t.Errorf("HashFile = %s, want %s (should match HashText of the same content)", got, want)
	}
}

func TestHashFile_NotFound(t *testing.T) {
	_, err := HashFile("sha256", filepath.Join(t.TempDir(), "does-not-exist.bin"), false)
	if err == nil {
		t.Fatal("expected an error for a missing file, got nil")
	}
	if CodeOf(err) != CodeNotFound {
		t.Errorf("CodeOf(err) = %d, want CodeNotFound (%d)", CodeOf(err), CodeNotFound)
	}
}

func TestHashFile_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.bin")
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	got, err := HashFile("sha256", path, false)
	if err != nil {
		t.Fatalf("unexpected error hashing an empty file: %v", err)
	}
	want := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if got != want {
		t.Errorf("hash of empty file = %s, want %s (sha256 of zero bytes)", got, want)
	}
}
