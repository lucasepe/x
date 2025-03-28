package template

import (
	"bytes"
	"io"
)

type EvalOptions struct {
	StartTag string
	EndTag   string
	Data     map[string]any
}

// Eval calls opts.Func on each template tag (placeholder) occurrence
// and substitutes it with the data written to TagFunc's writer.
//
// Returns the resulting slice of bytes that will be empty on error.
func Eval(template string, opts EvalOptions) ([]byte, error) {
	tpl := []byte(template)
	stg := []byte(opts.StartTag)
	etg := []byte(opts.EndTag)

	if n := bytes.Index(tpl, stg); n < 0 {
		return tpl, nil
	}

	bb := bytes.Buffer{}
	_, err := EvalTagFunc(&bb, EvalTagFuncOptions{
		Template: tpl,
		StartTag: stg,
		EndTag:   etg,
		Func:     KeepUnknownTags(stg, etg, opts.Data),
	})
	if err != nil {
		return bb.Bytes(), err
	}

	return bb.Bytes(), nil
}

type EvalTagFuncOptions struct {
	Template []byte
	StartTag []byte
	EndTag   []byte
	Func     TagFunc
}

// EvalTagFunc calls f on each template tag (placeholder) occurrence.
// Returns the number of bytes written to w.
func EvalTagFunc(wri io.Writer, opts EvalTagFuncOptions) (int64, error) {
	tpl := make([]byte, len(opts.Template))
	copy(tpl, opts.Template)

	stg := make([]byte, len(opts.StartTag))
	copy(stg, opts.StartTag)

	etg := make([]byte, len(opts.EndTag))
	copy(etg, opts.EndTag)

	var nn int64
	var ni int
	var err error
	for {
		n := bytes.Index(tpl, stg)
		if n < 0 {
			break
		}
		ni, err = wri.Write(tpl[:n])
		nn += int64(ni)
		if err != nil {
			return nn, err
		}

		tpl = tpl[n+len(stg):]
		n = bytes.Index(tpl, etg)
		if n < 0 {
			// cannot find end tag - just write it to the output.
			ni, _ = wri.Write(stg)
			nn += int64(ni)
			break
		}

		tag := bytes.TrimSpace(tpl[:n])
		ni, err = opts.Func(wri, tag)
		nn += int64(ni)
		if err != nil {
			return nn, err
		}
		tpl = tpl[n+len(etg):]
	}
	ni, err = wri.Write(tpl)
	nn += int64(ni)

	return nn, err
}
