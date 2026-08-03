package core

// sniffMIME identifies common image formats (and a couple of other
// binary types) by magic bytes. This intentionally covers only what
// the base64-image workflow needs — it is not a general-purpose MIME
// sniffer, and that's the point: net/http.DetectContentType would pull
// the entire net/http package (and its crypto/tls dependency) into
// even the CLI-only build. A dozen prefix checks do the actual job.
func sniffMIME(data []byte) string {
	switch {
	case hasPrefix(data, "\x89PNG\r\n\x1a\n"):
		return "image/png"
	case len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF:
		return "image/jpeg"
	case hasPrefix(data, "GIF87a") || hasPrefix(data, "GIF89a"):
		return "image/gif"
	case len(data) >= 12 && hasPrefix(data, "RIFF") && string(data[8:12]) == "WEBP":
		return "image/webp"
	case len(data) >= 2 && data[0] == 'B' && data[1] == 'M':
		return "image/bmp"
	case hasPrefix(data, "<svg") || hasPrefix(data, "<?xml"):
		return "image/svg+xml"
	case hasPrefix(data, "%PDF"):
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}

func hasPrefix(data []byte, prefix string) bool {
	if len(data) < len(prefix) {
		return false
	}
	return string(data[:len(prefix)]) == prefix
}
