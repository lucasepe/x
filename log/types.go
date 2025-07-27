package log

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
)

type ObjectEncoder interface {
	Encode(buf *bytes.Buffer)
}

type Field struct {
	Key     string
	Encoder ObjectEncoder
}

type stringVal string

func (s stringVal) Encode(buf *bytes.Buffer) {
	buf.WriteString(string(s))
}

type intVal int

func (i intVal) Encode(buf *bytes.Buffer) {
	buf.WriteString(strconv.Itoa(int(i)))
}

type boolVal bool

func (b boolVal) Encode(buf *bytes.Buffer) {
	if b {
		buf.WriteString("true")
	} else {
		buf.WriteString("false")
	}
}

type floatVal float64

func (f floatVal) Encode(buf *bytes.Buffer) {
	buf.WriteString(strconv.FormatFloat(float64(f), 'f', -1, 64))
}

type stringSliceVal []string

func (s stringSliceVal) Encode(buf *bytes.Buffer) {
	buf.WriteByte('[')
	for i, v := range s {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(v)
	}
	buf.WriteByte(']')
}

type mapStringAny map[string]any

// Encode produce una stringa tipo:
// chiave1=valore1 chiave2="valore 2" chiave3=nil
func (m mapStringAny) Encode(buf *bytes.Buffer) {
	buf.WriteByte('{')

	if len(m) == 0 {
		buf.WriteString("<empty>")
		buf.WriteByte('}')
		return
	}

	// Ordina le chiavi per output stabile
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, k := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		val := m[k]
		buf.WriteString(k)
		buf.WriteByte('=')

		switch v := val.(type) {
		case nil:
			buf.WriteString("nil")
		case string:
			// se contiene spazi, metti tra virgolette
			if len(v) > 0 && (containsSpace(v) || containsSpecial(v)) {
				fmt.Fprintf(buf, "%q", v)
			} else {
				buf.WriteString(v)
			}
		default:
			// per altri tipi usa %v
			fmt.Fprintf(buf, "%v", v)
		}
	}

	buf.WriteByte('}')
}

func containsSpace(s string) bool {
	for _, r := range s {
		if r == ' ' {
			return true
		}
	}
	return false
}

type errorVal struct {
	err error
}

func (e errorVal) Encode(buf *bytes.Buffer) {
	if e.err != nil {
		buf.WriteString(e.err.Error())
	}
}

func containsSpecial(s string) bool {
	for _, r := range s {
		if r == '=' || r == '"' {
			return true
		}
	}
	return false
}
