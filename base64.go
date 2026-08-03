package core

import (
	"encoding/base64"
	"os"
	"strings"
)

func b64Codec(urlSafe bool) *base64.Encoding {
	if urlSafe {
		return base64.URLEncoding
	}
	return base64.StdEncoding
}

func Base64EncodeText(text string, urlSafe bool) string {
	return b64Codec(urlSafe).EncodeToString([]byte(text))
}

func Base64DecodeText(text string, urlSafe bool) (string, error) {
	b, err := b64Codec(urlSafe).DecodeString(strings.TrimSpace(text))
	if err != nil {
		return "", NewInputError("invalid base64 input: " + err.Error())
	}
	return string(b), nil
}

// Base64EncodeFile reads a file (image or otherwise) and returns its
// base64 encoding, optionally wrapped as a data: URI. MIME type is
// detected from the first bytes via a tiny built-in magic-number
// sniffer (see mime.go) rather than net/http.DetectContentType, so the
// CLI-only build (see server_stub.go) never links the net/http stack.
func Base64EncodeFile(path string, asDataURI bool) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", NewNotFoundError("file not found: " + path)
		}
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	if asDataURI {
		return "data:" + sniffMIME(data) + ";base64," + encoded, nil
	}
	return encoded, nil
}

// Base64DecodeToFile decodes base64 (optionally a data: URI, whose
// header is stripped automatically) and writes the raw bytes to
// outPath, which is the safe way to round-trip binary content such as
// images through the CLI.
func Base64DecodeToFile(text, outPath string) error {
	text = stripDataURIPrefix(strings.TrimSpace(text))
	b, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return NewInputError("invalid base64 input: " + err.Error())
	}
	return os.WriteFile(outPath, b, 0o644)
}

func stripDataURIPrefix(s string) string {
	if strings.HasPrefix(s, "data:") {
		if idx := strings.Index(s, ","); idx != -1 {
			return s[idx+1:]
		}
	}
	return s
}
