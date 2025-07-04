package transport

import (
	"net/http"
	"testing"
)

func TestCloneRequest(t *testing.T) {
	orig := &http.Request{
		Header: http.Header{
			"Test-Header": []string{"value1", "value2"},
		},
	}
	clone := cloneRequest(orig)

	if &orig == &clone {
		t.Errorf("CloneRequest did not create a new request instance")
	}
	if &orig.Header == &clone.Header {
		t.Errorf("CloneRequest did not create a deep copy of headers")
	}
	if len(clone.Header.Get("Test-Header")) == 0 {
		t.Errorf("CloneRequest did not copy headers correctly")
	}
}

func TestCloneHeader(t *testing.T) {
	orig := http.Header{
		"Test-Header": []string{"value1", "value2"},
	}
	clone := cloneHeader(orig)

	if orig.Get("Test-Header") != clone.Get("Test-Header") {
		t.Errorf("cloneHeader did not create a deep copy of header values")
	}
}
