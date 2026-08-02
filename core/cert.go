package core

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
)

type CertInfo struct {
	Subject            string   `json:"subject"`
	Issuer             string   `json:"issuer"`
	SerialNumber       string   `json:"serial_number"`
	NotBefore          string   `json:"not_before"`
	NotAfter           string   `json:"not_after"`
	DNSNames           []string `json:"dns_names,omitempty"`
	EmailAddresses     []string `json:"email_addresses,omitempty"`
	IPAddresses        []string `json:"ip_addresses,omitempty"`
	SignatureAlgorithm string   `json:"signature_algorithm"`
	PublicKeyAlgorithm string   `json:"public_key_algorithm"`
	IsCA               bool     `json:"is_ca"`
}

// DecodeCertificate parses a PEM-wrapped or raw DER-encoded X.509
// certificate.
func DecodeCertificate(data []byte) (*CertInfo, error) {
	der := data
	if block, _ := pem.Decode(data); block != nil {
		der = block.Bytes
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, NewInputError("invalid certificate: " + err.Error())
	}

	ips := make([]string, 0, len(cert.IPAddresses))
	for _, ip := range cert.IPAddresses {
		ips = append(ips, ip.String())
	}

	return &CertInfo{
		Subject:            cert.Subject.String(),
		Issuer:             cert.Issuer.String(),
		SerialNumber:       cert.SerialNumber.String(),
		NotBefore:          cert.NotBefore.Format("2006-01-02 15:04:05 MST"),
		NotAfter:           cert.NotAfter.Format("2006-01-02 15:04:05 MST"),
		DNSNames:           cert.DNSNames,
		EmailAddresses:     cert.EmailAddresses,
		IPAddresses:        ips,
		SignatureAlgorithm: cert.SignatureAlgorithm.String(),
		PublicKeyAlgorithm: cert.PublicKeyAlgorithm.String(),
		IsCA:               cert.IsCA,
	}, nil
}

func DecodeCertificateFile(path string) (*CertInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, NewNotFoundError("file not found: " + path)
		}
		return nil, err
	}
	return DecodeCertificate(data)
}

// FormatCertInfo renders CertInfo as human-readable text for the CLI.
func FormatCertInfo(c *CertInfo) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Subject:             %s\n", c.Subject)
	fmt.Fprintf(&b, "Issuer:              %s\n", c.Issuer)
	fmt.Fprintf(&b, "Serial Number:       %s\n", c.SerialNumber)
	fmt.Fprintf(&b, "Not Before:          %s\n", c.NotBefore)
	fmt.Fprintf(&b, "Not After:           %s\n", c.NotAfter)
	fmt.Fprintf(&b, "Signature Algorithm: %s\n", c.SignatureAlgorithm)
	fmt.Fprintf(&b, "Public Key Algo:     %s\n", c.PublicKeyAlgorithm)
	fmt.Fprintf(&b, "Is CA:               %v\n", c.IsCA)
	if len(c.DNSNames) > 0 {
		fmt.Fprintf(&b, "DNS Names:           %s\n", strings.Join(c.DNSNames, ", "))
	}
	if len(c.EmailAddresses) > 0 {
		fmt.Fprintf(&b, "Email Addresses:     %s\n", strings.Join(c.EmailAddresses, ", "))
	}
	if len(c.IPAddresses) > 0 {
		fmt.Fprintf(&b, "IP Addresses:        %s\n", strings.Join(c.IPAddresses, ", "))
	}
	return b.String()
}
