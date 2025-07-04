package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
)

func VerboseRoundTripper(next http.RoundTripper) http.RoundTripper {
	return &verboseRoundTripper{
		next: next,
	}
}

type verboseRoundTripper struct {
	next http.RoundTripper
}

func (vt *verboseRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var reqBody []byte
	if req.Body != nil {
		reqBody, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
	}

	fmt.Fprintln(os.Stderr)

	dumpReq, _ := httputil.DumpRequestOut(req, false)
	addPrefixToLines(os.Stderr, dumpReq, "> ")

	if len(reqBody) > 0 {
		fmt.Fprintf(os.Stderr, "\n%s\n", string(reqBody))
	}
	fmt.Fprint(os.Stderr, "\n\n")

	resp, err := vt.next.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	var respBody []byte
	if resp.Body != nil {
		respBody, _ = io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewReader(respBody))
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		sampleSize := min(len(respBody), 512)
		contentType = http.DetectContentType(respBody[:sampleSize])
	}

	if isPrintable(contentType) {
		dumpResp, _ := httputil.DumpResponse(resp, false)
		addPrefixToLines(os.Stderr, dumpResp, "< ")

		fmt.Fprintln(os.Stderr)
		if strings.Contains(contentType, "application/json") {
			prettyPrintJSON(respBody)
		} else {
			fmt.Fprintln(os.Stderr, string(respBody))
		}
	}

	return resp, nil
}

func prettyPrintJSON(body []byte) {
	var out bytes.Buffer
	err := json.Indent(&out, body, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, string(body))
		return
	}
	fmt.Fprintln(os.Stderr, out.String())
}

func addPrefixToLines(w io.Writer, data []byte, prefix string) {
	lines := bytes.SplitSeq(data, []byte("\n"))
	for line := range lines {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		fmt.Fprintf(w, "%s%s\n", prefix, line)
	}
}

func isPrintable(contentType string) bool {
	ct := strings.ToLower(contentType)

	// Qualsiasi tipo text/*
	if strings.HasPrefix(ct, "text/") {
		return true
	}

	// Tipi MIME comuni per formati testuali e strutturati
	textTypes := []string{
		"application/json",
		"application/javascript",
		"application/xml",
		"application/x-yaml",
		"application/yaml",
		"application/toml",
		"application/x-toml",
		"application/x-www-form-urlencoded",
	}

	for _, t := range textTypes {
		if strings.Contains(ct, t) {
			return true
		}
	}

	return false
}
