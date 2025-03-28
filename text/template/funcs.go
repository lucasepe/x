package template

import (
	"io"
)

// TagFunc can be used as a substitution value in the map passed to Execute.
// * Must be safe to call from concurrently running goroutines.
// * Must write contents to w and return the number of bytes written.
type TagFunc func(w io.Writer, tag []byte) (int, error)

func KeepUnknownTags(startTag, endTag []byte, m map[string]any) TagFunc {
	return func(w io.Writer, tag []byte) (int, error) {
		v, ok := m[string(tag)]
		if !ok {
			if _, err := w.Write(startTag); err != nil {
				return 0, err
			}
			if _, err := w.Write(tag); err != nil {
				return 0, err
			}
			if _, err := w.Write(endTag); err != nil {
				return 0, err
			}
			return len(startTag) + len(tag) + len(endTag), nil
		}

		if v == nil {
			return 0, nil
		}

		switch value := v.(type) {
		case []byte:
			return w.Write(value)
		case string:
			return w.Write([]byte(value))
		case TagFunc:
			return value(w, tag)
		default:
			return w.Write([]byte(strval(value)))
		}
	}
}

// CollectTags accumulates all tags in the specified array.
func CollectTags(out *[]string) TagFunc {
	return func(_ io.Writer, tag []byte) (int, error) {
		*out = append(*out, string(tag))
		return 0, nil
	}
}
