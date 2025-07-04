package response_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/lucasepe/x/http/response"
)

func TestDumpResponse(t *testing.T) {
	tests := []struct {
		name       string
		resp       *http.Response
		wantOut    string
		wantErrOut string
		wantErr    bool
	}{
		{
			name: "200 OK writes to outwri",
			resp: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("success!")),
			},
			wantOut:    "success!",
			wantErrOut: "",
			wantErr:    false,
		},
		{
			name: "404 writes to errwri and returns error",
			resp: &http.Response{
				StatusCode: 404,
				Body:       io.NopCloser(strings.NewReader("not found")),
			},
			wantOut:    "",
			wantErrOut: "not found",
			wantErr:    true,
		},
		{
			name: "204 No Content with nil body (still OK)",
			resp: &http.Response{
				StatusCode: 204,
				Body:       nil,
			},
			wantOut:    "",
			wantErrOut: "",
			wantErr:    false,
		},
		{
			name: "500 with nil body returns error but no panic",
			resp: &http.Response{
				StatusCode: 500,
				Body:       nil,
			},
			wantOut:    "",
			wantErrOut: "",
			wantErr:    true,
		},
		{
			name:    "nil response returns error",
			resp:    nil,
			wantErr: true,
		},
		{
			name: "error copying body returns wrapped error",
			resp: &http.Response{
				StatusCode: 200,
				Body:       &fakeBody{},
			},
			wantOut:    "",
			wantErrOut: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var outBuf, errBuf bytes.Buffer

			err := response.Dump(tt.resp, &outBuf, &errBuf)

			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error: got %v, wantErr %v", err, tt.wantErr)
			}

			if got := outBuf.String(); got != tt.wantOut {
				t.Errorf("unexpected output to outwri: got %q, want %q", got, tt.wantOut)
			}

			if got := errBuf.String(); got != tt.wantErrOut {
				t.Errorf("unexpected output to errwri: got %q, want %q", got, tt.wantErrOut)
			}
		})
	}
}

// fakeBody simula un body che fallisce alla lettura, per testare gli errori
type fakeBody struct{}

func (f *fakeBody) Read([]byte) (int, error) { return 0, errors.New("read error") }
func (f *fakeBody) Close() error             { return nil }
