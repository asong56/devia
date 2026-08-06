//go:build !noserve

// This file is excluded from the "noserve" build (see stub.go), which
// produces devia-cli: a smaller binary for people who only want the
// command line and never touch net/http at all. Everything here calls
// straight into the same core package the CLI uses — the API is a
// thin adapter, not a second implementation.
package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"devia/internal/core"
	"devia/internal/version"
)

// Run starts the JSON API and blocks until it exits (normally only on
// error, since http.ListenAndServe never returns nil).
func Run(host string, port int) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handleIndex)
	mux.HandleFunc("GET /api/v1/health", handleHealth)
	mux.HandleFunc("POST /api/v1/hash", handleHash)
	mux.HandleFunc("POST /api/v1/checksum", handleChecksum)
	mux.HandleFunc("POST /api/v1/base64/encode", handleBase64Encode)
	mux.HandleFunc("POST /api/v1/base64/decode", handleBase64Decode)
	mux.HandleFunc("POST /api/v1/json/format", handleJSONFormat)
	mux.HandleFunc("POST /api/v1/json/minify", handleJSONMinify)
	mux.HandleFunc("POST /api/v1/escape", handleEscape)
	mux.HandleFunc("POST /api/v1/unescape", handleUnescape)
	mux.HandleFunc("POST /api/v1/uuid", handleUUID)
	mux.HandleFunc("POST /api/v1/text", handleText)
	mux.HandleFunc("POST /api/v1/lorem", handleLorem)
	mux.HandleFunc("POST /api/v1/timestamp", handleTimestamp)
	mux.HandleFunc("POST /api/v1/radix", handleRadix)
	mux.HandleFunc("POST /api/v1/cron", handleCron)
	mux.HandleFunc("POST /api/v1/regex/test", handleRegexTest)
	mux.HandleFunc("POST /api/v1/regex/replace", handleRegexReplace)
	mux.HandleFunc("POST /api/v1/diff", handleDiff)
	mux.HandleFunc("POST /api/v1/cert/decode", handleCertDecode)

	addr := fmt.Sprintf("%s:%d", host, port)
	fmt.Fprintf(os.Stderr, "devia: API listening on http://%s (Ctrl+C to stop)\n", addr)
	fmt.Fprintf(os.Stderr, "devia: docs at            http://%s/\n", addr)
	return http.ListenAndServe(addr, mux)
}

// ---- response envelope: identical shape to the CLI's --json mode ----

type apiEnvelope struct {
	OK     bool        `json:"ok"`
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
	Code   int         `json:"code,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeOK(w http.ResponseWriter, result interface{}) {
	writeJSON(w, http.StatusOK, apiEnvelope{OK: true, Result: result})
}

// writeErr maps devia's standard error codes onto proper HTTP status
// codes, so a caller can branch on the status alone without parsing
// the body — the JSON body still carries the same {ok,error,code} for
// callers who want it.
func writeErr(w http.ResponseWriter, err error) {
	code := core.CodeOf(err)
	status := http.StatusInternalServerError
	switch code {
	case core.CodeInput, core.CodeUsage:
		status = http.StatusBadRequest
	case core.CodeNotFound:
		status = http.StatusNotFound
	}
	writeJSON(w, status, apiEnvelope{OK: false, Error: err.Error(), Code: code})
}

func decodeBody(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return core.NewInputError("invalid JSON request body: " + err.Error())
	}
	return nil
}

// ---- handlers ----

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(indexHTML))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeOK(w, map[string]string{"name": "devia", "version": version.Version})
}

func handleHash(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text   string `json:"text"`
		Algo   string `json:"algo"`
		HMAC   string `json:"hmac"`
		Base64 bool   `json:"base64"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	result, err := core.HashText(req.Algo, req.Text, req.HMAC, req.Base64)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeOK(w, result)
}

// handleChecksum accepts base64-encoded file content in the JSON body
// (there's no multipart upload here on purpose — keeping the API
// pure-JSON is what keeps this file small and dependency-free).
func handleChecksum(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Data    string `json:"data"` // base64-encoded bytes
		Algo    string `json:"algo"`
		Compare string `json:"compare"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	raw, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		writeErr(w, core.NewInputError("invalid base64 in 'data' field"))
		return
	}
	sum, err := core.HashBytes(req.Algo, raw, false)
	if err != nil {
		writeErr(w, err)
		return
	}
	if req.Compare != "" {
		writeOK(w, map[string]interface{}{
			"checksum": sum,
			"match":    strings.EqualFold(strings.TrimSpace(sum), strings.TrimSpace(req.Compare)),
		})
		return
	}
	writeOK(w, sum)
}

func handleBase64Encode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text string `json:"text"`
		URL  bool   `json:"url"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	writeOK(w, core.Base64EncodeText(req.Text, req.URL))
}

func handleBase64Decode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text string `json:"text"`
		URL  bool   `json:"url"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	result, err := core.Base64DecodeText(req.Text, req.URL)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeOK(w, result)
}

func handleJSONFormat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text   string `json:"text"`
		Indent string `json:"indent"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	result, err := core.JSONFormat(req.Text, req.Indent)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeOK(w, result)
}

func handleJSONMinify(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text string `json:"text"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	result, err := core.JSONMinify(req.Text)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeOK(w, result)
}

func handleEscape(w http.ResponseWriter, r *http.Request)   { escapeAPI(w, r, false) }
func handleUnescape(w http.ResponseWriter, r *http.Request) { escapeAPI(w, r, true) }

func escapeAPI(w http.ResponseWriter, r *http.Request, unescape bool) {
	var req struct {
		Mode string `json:"mode"`
		Text string `json:"text"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	var result string
	var err error
	switch req.Mode {
	case "json":
		if unescape {
			result, err = core.UnescapeJSON(req.Text)
		} else {
			result, err = core.EscapeJSON(req.Text)
		}
	case "url":
		if unescape {
			result, err = core.UnescapeURL(req.Text, false)
		} else {
			result = core.EscapeURL(req.Text, false)
		}
	case "url-path":
		if unescape {
			result, err = core.UnescapeURL(req.Text, true)
		} else {
			result = core.EscapeURL(req.Text, true)
		}
	case "html":
		if unescape {
			result = core.UnescapeHTML(req.Text)
		} else {
			result = core.EscapeHTML(req.Text)
		}
	case "unicode":
		if unescape {
			result, err = core.UnescapeUnicode(req.Text)
		} else {
			result = core.EscapeUnicode(req.Text)
		}
	default:
		err = core.NewInputError("unknown mode: " + req.Mode)
	}
	if err != nil {
		writeErr(w, err)
		return
	}
	writeOK(w, result)
}

func handleUUID(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Count int  `json:"count"`
		Upper bool `json:"upper"`
	}
	// Body is optional here (a bare POST with no body is a valid way to
	// ask for one UUID), so a decode error is not fatal — just fall
	// back to defaults.
	_ = decodeBody(r, &req)
	if req.Count == 0 {
		req.Count = 1
	}
	ids, err := core.NewUUIDs(req.Count, req.Upper)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeOK(w, ids)
}

func handleText(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text string `json:"text"`
		Mode string `json:"mode"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	result, err := core.TextTransform(req.Mode, req.Text)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeOK(w, result)
}

func handleLorem(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type    string `json:"type"`
		Count   int    `json:"count"`
		Classic bool   `json:"classic"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	result, err := core.LoremGenerate(req.Type, req.Count, req.Classic)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeOK(w, result)
}

func handleTimestamp(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Action    string `json:"action"` // now|to-date|from-date
		Timestamp int64  `json:"timestamp"`
		Date      string `json:"date"`
		TZ        string `json:"tz"`
		Format    string `json:"format"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	switch req.Action {
	case "now", "":
		writeOK(w, core.NowUnix())
	case "to-date":
		result, err := core.UnixToDate(req.Timestamp, req.TZ, req.Format)
		if err != nil {
			writeErr(w, err)
			return
		}
		writeOK(w, result)
	case "from-date":
		result, err := core.DateToUnix(req.Date, req.TZ, req.Format)
		if err != nil {
			writeErr(w, err)
			return
		}
		writeOK(w, result)
	default:
		writeErr(w, core.NewInputError("unknown action: "+req.Action))
	}
}

func handleRadix(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Value string `json:"value"`
		From  int    `json:"from"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	result, err := core.ConvertRadix(req.Value, req.From)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeOK(w, result)
}

func handleCron(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Expr string `json:"expr"`
		Next int    `json:"next"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	if req.Next == 0 {
		req.Next = 5
	}
	spec, err := core.ParseCron(req.Expr)
	if err != nil {
		writeErr(w, err)
		return
	}
	times, err := spec.Next(time.Now(), req.Next)
	if err != nil {
		writeErr(w, err)
		return
	}
	out := make([]string, len(times))
	for i, t := range times {
		out[i] = t.Format("2006-01-02 15:04:05 Mon")
	}
	writeOK(w, map[string]interface{}{
		"description": core.DescribeCron(req.Expr),
		"next":        out,
	})
}

func handleRegexTest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Pattern string `json:"pattern"`
		Flags   string `json:"flags"`
		Text    string `json:"text"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	matches, err := core.RegexTest(req.Pattern, req.Flags, req.Text)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeOK(w, matches)
}

func handleRegexReplace(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Pattern     string `json:"pattern"`
		Flags       string `json:"flags"`
		Text        string `json:"text"`
		Replacement string `json:"replacement"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	result, err := core.RegexReplace(req.Pattern, req.Flags, req.Text, req.Replacement)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeOK(w, result)
}

func handleDiff(w http.ResponseWriter, r *http.Request) {
	var req struct {
		A string `json:"a"`
		B string `json:"b"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	lines, err := core.DiffText(req.A, req.B)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeOK(w, lines)
}

// handleCertDecode accepts either raw PEM text or base64-encoded
// DER/PEM bytes, so both `curl -d '{"pem":"-----BEGIN..."}'` and a
// programmatic base64 upload work.
func handleCertDecode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PEM  string `json:"pem"`
		Data string `json:"data"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, err)
		return
	}
	var raw []byte
	switch {
	case req.PEM != "":
		raw = []byte(req.PEM)
	case req.Data != "":
		b, err := base64.StdEncoding.DecodeString(req.Data)
		if err != nil {
			writeErr(w, core.NewInputError("invalid base64 in 'data' field"))
			return
		}
		raw = b
	default:
		writeErr(w, core.NewInputError("provide either 'pem' (raw text) or 'data' (base64)"))
		return
	}
	info, err := core.DecodeCertificate(raw)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeOK(w, info)
}

const indexHTML = `<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>devia API</title>
<style>
body{font-family:ui-monospace,Consolas,monospace;max-width:720px;margin:40px auto;padding:0 16px;color:#222}
code{background:#eee;padding:2px 6px;border-radius:3px}
table{width:100%;border-collapse:collapse;font-size:14px}
td{padding:6px 10px;border-bottom:1px solid #ddd;vertical-align:top}
td:first-child{color:#888;white-space:nowrap}
</style></head><body>
<h2>devia API</h2>
<p>POST a JSON body to any endpoint below. Every response is
<code>{"ok":true,"result":...}</code> (HTTP 200) or
<code>{"ok":false,"error":"...","code":N}</code> (HTTP 400/404/500).</p>
<table>
<tr><td>POST</td><td>/api/v1/hash</td><td>{text, algo, hmac, base64}</td></tr>
<tr><td>POST</td><td>/api/v1/checksum</td><td>{data (base64), algo, compare}</td></tr>
<tr><td>POST</td><td>/api/v1/base64/encode</td><td>{text, url}</td></tr>
<tr><td>POST</td><td>/api/v1/base64/decode</td><td>{text, url}</td></tr>
<tr><td>POST</td><td>/api/v1/json/format</td><td>{text, indent}</td></tr>
<tr><td>POST</td><td>/api/v1/json/minify</td><td>{text}</td></tr>
<tr><td>POST</td><td>/api/v1/escape</td><td>{mode, text}</td></tr>
<tr><td>POST</td><td>/api/v1/unescape</td><td>{mode, text}</td></tr>
<tr><td>POST</td><td>/api/v1/uuid</td><td>{count, upper}</td></tr>
<tr><td>POST</td><td>/api/v1/text</td><td>{text, mode}</td></tr>
<tr><td>POST</td><td>/api/v1/lorem</td><td>{type, count, classic}</td></tr>
<tr><td>POST</td><td>/api/v1/timestamp</td><td>{action, timestamp, date, tz, format}</td></tr>
<tr><td>POST</td><td>/api/v1/radix</td><td>{value, from}</td></tr>
<tr><td>POST</td><td>/api/v1/cron</td><td>{expr, next}</td></tr>
<tr><td>POST</td><td>/api/v1/regex/test</td><td>{pattern, flags, text}</td></tr>
<tr><td>POST</td><td>/api/v1/regex/replace</td><td>{pattern, flags, text, replacement}</td></tr>
<tr><td>POST</td><td>/api/v1/diff</td><td>{a, b}</td></tr>
<tr><td>POST</td><td>/api/v1/cert/decode</td><td>{pem} or {data (base64)}</td></tr>
<tr><td>GET</td><td>/api/v1/health</td><td>-</td></tr>
</table>
</body></html>`
