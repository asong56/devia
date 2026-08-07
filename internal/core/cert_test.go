package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testCertPEM is a real, openssl-generated self-signed RSA-2048
// certificate (not synthetic bytes) with a full RFC 2253-ish Subject/
// Issuer DN and a SAN extension covering DNS, IP, and email names —
// deliberately exercising every field CertInfo populates. It was
// generated once with:
//
//	openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem \
//	  -days 3650 -nodes \
//	  -subj "/C=US/ST=California/L=San Francisco/O=Devia Test/CN=devia.example.com" \
//	  -addext "subjectAltName=DNS:devia.example.com,DNS:www.devia.example.com,IP:127.0.0.1,email:admin@devia.example.com"
//
// and is now frozen as a fixture: every field below (serial number,
// validity window, DN order) was cross-checked against `openssl x509
// -text` output for this exact certificate, so it stays valid forever
// regardless of when `go test` actually runs.
const testCertPEM = `-----BEGIN CERTIFICATE-----
MIIEDTCCAvWgAwIBAgIUXHlHfJLjHBcFqDzuIhJQMQQ3MqgwDQYJKoZIhvcNAQEL
BQAwazELMAkGA1UEBhMCVVMxEzARBgNVBAgMCkNhbGlmb3JuaWExFjAUBgNVBAcM
DVNhbiBGcmFuY2lzY28xEzARBgNVBAoMCkRldmlhIFRlc3QxGjAYBgNVBAMMEWRl
dmlhLmV4YW1wbGUuY29tMB4XDTI2MDgwNzAyMDEyMloXDTM2MDgwNDAyMDEyMlow
azELMAkGA1UEBhMCVVMxEzARBgNVBAgMCkNhbGlmb3JuaWExFjAUBgNVBAcMDVNh
biBGcmFuY2lzY28xEzARBgNVBAoMCkRldmlhIFRlc3QxGjAYBgNVBAMMEWRldmlh
LmV4YW1wbGUuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA4S1N
/ijkw49HDcRiOfYmDUVOIA7jAMrmELK01mMNZ8FNrlkw1KAShZMRwv5MfRmuAbR0
+CkgUeaMrCWxJBhvFiyA5LEZxXXAZXefcs0aGVTKI5FyiD87hCU33NyUk0eOtv+W
hJ9DSDzA8fVb3VUDxrmI7i01r329IAcSzvvY1ZbBmVylbN3J6EZ5MSK4wGxzm4Az
vGNWhPeVMY6tHkTfJInLWq4B7B8H/COVMn+cNi7dHhPj02DSc54VDK4HVcw624mL
q8DKMEthBshOStWFTBX+Bp3ltI399km1G4QhqA1ilr3Ad89SlM/8Tc4+WgZG8yrt
qIDq73HXMyQCMqqbowIDAQABo4GoMIGlMB0GA1UdDgQWBBQ5JLsMHUHzOjp6SbsF
qx2b84mAXDAfBgNVHSMEGDAWgBQ5JLsMHUHzOjp6SbsFqx2b84mAXDAPBgNVHRMB
Af8EBTADAQH/MFIGA1UdEQRLMEmCEWRldmlhLmV4YW1wbGUuY29tghV3d3cuZGV2
aWEuZXhhbXBsZS5jb22HBH8AAAGBF2FkbWluQGRldmlhLmV4YW1wbGUuY29tMA0G
CSqGSIb3DQEBCwUAA4IBAQBFg3W0rH5UJmdf687nhMhdVnhFmulyvud5mEFwBDy6
s3fvUn1qCJrfq1j/HHHV3JL2esslBbbjOYD000Dazwa95gpsMmJDRRlEZGD69j1j
rAnRIOm6ge0/PZ3HYIULWOLWuK4u+aJE4mcFVzsPNo+9L3MAWt/igG+f8LnLbSVN
eGIASO/d41qlJhQYr00nNALd2jxUn4oGtg/NnfVSolWSoemUBqHtyRlUQs1J1MnN
1FVGzQxYRbEf59bpOVDs/a/B+zHFNhGuGw5zXsHzMZLm37qg4/Lg031S28XYdkOP
qiszjC+kq1yqmRhTDPsYEiy91E6oHwvJb40N7SRRdLuU
-----END CERTIFICATE-----
`

func TestDecodeCertificate_AllFields(t *testing.T) {
	info, err := DecodeCertificate([]byte(testCertPEM))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Subject/Issuer via pkix.Name.String(): the actual field order
	// produced by Go's pkix package for this certificate is
	// CN, O, L, ST, C — not the RFC 2253 reverse-of-input order
	// that the documentation implies. This expected value was
	// confirmed by running the test and reading the actual output.
	wantDN := "CN=devia.example.com,O=Devia Test,L=San Francisco,ST=California,C=US"
	if info.Subject != wantDN {
		t.Errorf("Subject = %q, want %q", info.Subject, wantDN)
	}
	if info.Issuer != wantDN {
		t.Errorf("Issuer = %q, want %q (self-signed: issuer == subject)", info.Issuer, wantDN)
	}

	wantSerial := "527931768447376603306409014135775709616439505576"
	if info.SerialNumber != wantSerial {
		t.Errorf("SerialNumber = %q, want %q", info.SerialNumber, wantSerial)
	}

	if info.NotBefore != "2026-08-07 02:01:22 UTC" {
		t.Errorf("NotBefore = %q, want %q", info.NotBefore, "2026-08-07 02:01:22 UTC")
	}
	if info.NotAfter != "2036-08-04 02:01:22 UTC" {
		t.Errorf("NotAfter = %q, want %q", info.NotAfter, "2036-08-04 02:01:22 UTC")
	}

	wantDNS := []string{"devia.example.com", "www.devia.example.com"}
	if len(info.DNSNames) != len(wantDNS) || info.DNSNames[0] != wantDNS[0] || info.DNSNames[1] != wantDNS[1] {
		t.Errorf("DNSNames = %v, want %v", info.DNSNames, wantDNS)
	}

	wantEmails := []string{"admin@devia.example.com"}
	if len(info.EmailAddresses) != 1 || info.EmailAddresses[0] != wantEmails[0] {
		t.Errorf("EmailAddresses = %v, want %v", info.EmailAddresses, wantEmails)
	}

	wantIPs := []string{"127.0.0.1"}
	if len(info.IPAddresses) != 1 || info.IPAddresses[0] != wantIPs[0] {
		t.Errorf("IPAddresses = %v, want %v", info.IPAddresses, wantIPs)
	}

	if info.SignatureAlgorithm != "SHA256-RSA" {
		t.Errorf("SignatureAlgorithm = %q, want %q", info.SignatureAlgorithm, "SHA256-RSA")
	}
	if info.PublicKeyAlgorithm != "RSA" {
		t.Errorf("PublicKeyAlgorithm = %q, want %q", info.PublicKeyAlgorithm, "RSA")
	}
	if !info.IsCA {
		t.Error("IsCA = false, want true (this fixture has Basic Constraints CA:TRUE)")
	}
}

func TestDecodeCertificate_InvalidPEM(t *testing.T) {
	_, err := DecodeCertificate([]byte("not a certificate at all"))
	if err == nil {
		t.Fatal("expected an error for non-certificate input, got nil")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("CodeOf(err) = %d, want CodeInput (%d)", CodeOf(err), CodeInput)
	}
}

func TestDecodeCertificate_TruncatedPEM(t *testing.T) {
	// A PEM block that decodes but contains truncated/corrupt DER —
	// the "malformed input from a different tool" case: someone's
	// editor mangled a line, or a copy-paste dropped the last few
	// lines of a cert.
	truncated := testCertPEM[:len(testCertPEM)/2] + "\n-----END CERTIFICATE-----\n"
	_, err := DecodeCertificate([]byte(truncated))
	if err == nil {
		t.Fatal("expected an error for truncated certificate data, got nil")
	}
	if CodeOf(err) != CodeInput {
		t.Errorf("CodeOf(err) = %d, want CodeInput (%d)", CodeOf(err), CodeInput)
	}
}

func TestDecodeCertificateFile_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.crt")
	if err := os.WriteFile(path, []byte(testCertPEM), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	info, err := DecodeCertificateFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.SignatureAlgorithm != "SHA256-RSA" {
		t.Errorf("SignatureAlgorithm = %q, want %q", info.SignatureAlgorithm, "SHA256-RSA")
	}
}

func TestDecodeCertificateFile_NotFound(t *testing.T) {
	_, err := DecodeCertificateFile(filepath.Join(t.TempDir(), "nope.crt"))
	if err == nil {
		t.Fatal("expected an error for a missing file, got nil")
	}
	if CodeOf(err) != CodeNotFound {
		t.Errorf("CodeOf(err) = %d, want CodeNotFound (%d)", CodeOf(err), CodeNotFound)
	}
}

func TestFormatCertInfo_ContainsKeyFields(t *testing.T) {
	info, err := DecodeCertificate([]byte(testCertPEM))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := FormatCertInfo(info)
	for _, want := range []string{
		"Subject:",
		"devia.example.com",
		"Serial Number:",
		"527931768447376603306409014135775709616439505576",
		"SHA256-RSA",
		"Is CA:               true",
		"DNS Names:",
		"www.devia.example.com",
		"Email Addresses:",
		"admin@devia.example.com",
		"IP Addresses:",
		"127.0.0.1",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("FormatCertInfo output missing expected substring %q\nfull output:\n%s", want, out)
		}
	}
}
