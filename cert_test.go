package core

import (
	"path/filepath"
	"testing"
)

// Note: there's no positive-path test here with a real certificate.
// Hand-authoring a valid X.509 DER/PEM fixture without a working Go
// toolchain to generate and verify one isn't something we can do
// reliably, so this file sticks to the error paths, which don't
// require a real certificate.

func TestDecodeCertificateInvalidData(t *testing.T) {
	_, err := DecodeCertificate([]byte("this is not a certificate"))
	if err == nil {
		t.Fatal("expected an error for non-certificate data")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("invalid certificate data should be a CodeInput error, got code %d", CodeOf(err))
	}
}

func TestDecodeCertificateEmptyInput(t *testing.T) {
	_, err := DecodeCertificate([]byte(""))
	if err == nil {
		t.Fatal("expected an error for empty input")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("expected CodeInput, got code %d", CodeOf(err))
	}
}

func TestDecodeCertificateFileMissing(t *testing.T) {
	_, err := DecodeCertificateFile(filepath.Join(t.TempDir(), "nope.pem"))
	if err == nil {
		t.Fatal("expected an error for a missing file")
	}
	if CodeOf(err) != CodeNotFound {
		t.Errorf("missing file should be a CodeNotFound error, got code %d", CodeOf(err))
	}
}
