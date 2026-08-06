package core

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"hash"
	"io"
	"os"
)

func newHasher(algo string) (hash.Hash, error) {
	switch algo {
	case "md5":
		return md5.New(), nil
	case "sha1":
		return sha1.New(), nil
	case "sha256", "":
		return sha256.New(), nil
	case "sha512":
		return sha512.New(), nil
	default:
		return nil, NewInputError("unsupported algorithm: " + algo + " (want md5|sha1|sha256|sha512)")
	}
}

func newHMACHasher(algo, key string) (hash.Hash, error) {
	switch algo {
	case "md5":
		return hmac.New(md5.New, []byte(key)), nil
	case "sha1":
		return hmac.New(sha1.New, []byte(key)), nil
	case "sha256", "":
		return hmac.New(sha256.New, []byte(key)), nil
	case "sha512":
		return hmac.New(sha512.New, []byte(key)), nil
	default:
		return nil, NewInputError("unsupported algorithm: " + algo + " (want md5|sha1|sha256|sha512)")
	}
}

func encodeSum(sum []byte, useBase64 bool) string {
	if useBase64 {
		return base64.StdEncoding.EncodeToString(sum)
	}
	return hex.EncodeToString(sum)
}

// HashText hashes a UTF-8 string. If hmacKey is non-empty, an HMAC is
// computed instead of a plain hash.
func HashText(algo, text, hmacKey string, useBase64 bool) (string, error) {
	var h hash.Hash
	var err error
	if hmacKey != "" {
		h, err = newHMACHasher(algo, hmacKey)
	} else {
		h, err = newHasher(algo)
	}
	if err != nil {
		return "", err
	}
	h.Write([]byte(text))
	return encodeSum(h.Sum(nil), useBase64), nil
}

// HashBytes hashes an in-memory byte slice (used by the HTTP API, where
// file content arrives base64-encoded inside a JSON body).
func HashBytes(algo string, data []byte, useBase64 bool) (string, error) {
	h, err := newHasher(algo)
	if err != nil {
		return "", err
	}
	h.Write(data)
	return encodeSum(h.Sum(nil), useBase64), nil
}

// HashFile streams a file through the hasher without loading it fully
// into memory, so it stays cheap even for large files.
func HashFile(algo, path string, useBase64 bool) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", NewNotFoundError("file not found: " + path)
		}
		return "", err
	}
	defer f.Close()

	h, err := newHasher(algo)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return encodeSum(h.Sum(nil), useBase64), nil
}
