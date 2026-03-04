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

// DefaultVerboseMaxBodyBytes e' il numero massimo di byte di body stampati per
// richiesta o risposta quando il logging del body e' attivo.
const DefaultVerboseMaxBodyBytes = 8 * 1024

// VerboseOptions controlla il livello di dettaglio del transport di debug.
type VerboseOptions struct {
	LogBodies    bool
	MaxBodyBytes int
}

// VerboseRoundTripper stampa richiesta e risposta su stderr in forma leggibile.
// Per default logga anche il body, ma ne limita l'anteprima a 8 KiB per
// evitare di bufferizzare interi payload grandi in memoria.
func VerboseRoundTripper(next http.RoundTripper) http.RoundTripper {
	return VerboseRoundTripperWithOptions(next, VerboseOptions{
		LogBodies:    true,
		MaxBodyBytes: DefaultVerboseMaxBodyBytes,
	})
}

// VerboseRoundTripperWithOptions crea un transport di debug configurabile.
// Il body, se abilitato, viene campionato e reinserito nello stream senza
// consumare l'intera request o response in memoria.
func VerboseRoundTripperWithOptions(next http.RoundTripper, opts VerboseOptions) http.RoundTripper {
	if next == nil {
		next = Default()
	}
	if opts.MaxBodyBytes < 0 {
		opts.MaxBodyBytes = 0
	}

	return &verboseRoundTripper{
		next: next,
		opts: opts,
	}
}

type verboseRoundTripper struct {
	next http.RoundTripper
	opts VerboseOptions
}

func (vt *verboseRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	reqBody, reqBodyTruncated, err := previewRequestBody(req, vt.opts.MaxBodyBytes, vt.opts.LogBodies)
	if err != nil {
		return nil, err
	}

	fmt.Fprintln(os.Stderr)

	dumpReq, _ := httputil.DumpRequestOut(req, false)
	addPrefixToLines(os.Stderr, dumpReq, "> ")

	if vt.opts.LogBodies {
		printBodyPreview(reqBody, reqBodyTruncated, req.Header.Get("Content-Type"))
	}
	fmt.Fprint(os.Stderr, "\n\n")

	resp, err := vt.next.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	var respBody []byte
	var respBodyTruncated bool
	if vt.opts.LogBodies {
		respBody, respBodyTruncated, resp.Body, err = previewReadCloser(resp.Body, vt.opts.MaxBodyBytes)
		if err != nil {
			return nil, err
		}
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" && len(respBody) > 0 {
		sampleSize := min(len(respBody), 512)
		contentType = http.DetectContentType(respBody[:sampleSize])
	}

	if !vt.opts.LogBodies || isPrintable(contentType) {
		dumpResp, _ := httputil.DumpResponse(resp, false)
		addPrefixToLines(os.Stderr, dumpResp, "< ")

		if vt.opts.LogBodies {
			printBodyPreview(respBody, respBodyTruncated, contentType)
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

func printBodyPreview(body []byte, truncated bool, contentType string) {
	if len(body) == 0 {
		return
	}

	fmt.Fprintln(os.Stderr)
	if !truncated && strings.Contains(strings.ToLower(contentType), "application/json") {
		prettyPrintJSON(body)
	} else {
		fmt.Fprintln(os.Stderr, string(body))
	}
	if truncated {
		fmt.Fprintf(os.Stderr, "[body truncated to %d bytes]\n", len(body))
	}
}

func previewRequestBody(req *http.Request, maxBytes int, enabled bool) ([]byte, bool, error) {
	if !enabled || req == nil || req.Body == nil || req.Body == http.NoBody || maxBytes == 0 {
		return nil, false, nil
	}

	preview, truncated, restored, err := previewReadCloser(req.Body, maxBytes)
	if err != nil {
		return nil, false, err
	}
	req.Body = restored
	if req.GetBody != nil {
		copied := append([]byte(nil), preview...)
		if truncated {
			// Se il body supera il limite non possiamo ricostruire il payload
			// completo da questa sola preview, quindi lasciamo invariato GetBody.
			return preview, true, nil
		}
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(copied)), nil
		}
	}
	return preview, truncated, nil
}

func previewReadCloser(rc io.ReadCloser, maxBytes int) ([]byte, bool, io.ReadCloser, error) {
	if rc == nil || rc == http.NoBody || maxBytes == 0 {
		return nil, false, rc, nil
	}

	consumed, err := io.ReadAll(io.LimitReader(rc, int64(maxBytes+1)))
	if err != nil {
		return nil, false, rc, err
	}

	truncated := len(consumed) > maxBytes
	preview := consumed
	if truncated {
		preview = consumed[:maxBytes]
	}

	restored := &compositeReadCloser{
		Reader: io.MultiReader(bytes.NewReader(consumed), rc),
		Closer: rc,
	}

	return preview, truncated, restored, nil
}

type compositeReadCloser struct {
	io.Reader
	io.Closer
}

func isPrintable(contentType string) bool {
	ct := strings.ToLower(contentType)

	// Qualsiasi tipo text/* e' sicuro da stampare direttamente.
	if strings.HasPrefix(ct, "text/") {
		return true
	}

	// Tipi MIME comuni per formati testuali e strutturati.
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
