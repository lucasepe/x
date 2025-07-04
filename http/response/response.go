package response

import (
	"fmt"
	"io"
	"net/http"
)

func Dump(res *http.Response, outwri, errwri io.Writer) error {
	if res == nil {
		return fmt.Errorf("nil http.Response received")
	}

	if outwri == nil {
		outwri = io.Discard
	}
	if errwri == nil {
		errwri = io.Discard
	}

	statusOK := res.StatusCode >= 200 && res.StatusCode < 300

	if res.Body == nil {
		if !statusOK {
			return fmt.Errorf("http request failed with status: %d %s", res.StatusCode, http.StatusText(res.StatusCode))
		}
		// Success but no body: still OK
		return nil
	}
	defer res.Body.Close()

	if !statusOK {
		if _, err := io.Copy(errwri, res.Body); err != nil {
			return fmt.Errorf("http status %d; also failed to read body: %w", res.StatusCode, err)
		}
		return fmt.Errorf("http request failed with status: %d %s", res.StatusCode, http.StatusText(res.StatusCode))
	}

	if _, err := io.Copy(outwri, res.Body); err != nil {
		return fmt.Errorf("failed to read response body (status %d): %w", res.StatusCode, err)
	}
	return nil
}
