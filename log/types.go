package log

import (
	"bytes"
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

type mapStringVal map[string]string

func (m mapStringVal) Encode(buf *bytes.Buffer) {
	buf.WriteByte('{')
	i := 0
	for k, v := range m {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(k)
		buf.WriteByte(':')
		buf.WriteString(v)
		i++
	}
	buf.WriteByte('}')
}

type errorVal struct {
	err error
}

func (e errorVal) Encode(buf *bytes.Buffer) {
	if e.err != nil {
		buf.WriteString(e.err.Error())
	}
}
