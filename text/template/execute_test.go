package template

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvalTagFunc(t *testing.T) {
	tests := []struct {
		name     string
		template []byte
		startTag []byte
		endTag   []byte
		values   map[string]any
		want     string
	}{
		{
			name:     "single known tag",
			template: []byte("Hello {{name}}!"),
			startTag: []byte("{{"),
			endTag:   []byte("}}"),
			values:   map[string]any{"name": "World"},
			want:     "Hello World!",
		},
		{
			name:     "multiple known tags",
			template: []byte("{{greet}}, {{name}}!"),
			startTag: []byte("{{"),
			endTag:   []byte("}}"),
			values:   map[string]any{"greet": "Hi", "name": "Alice"},
			want:     "Hi, Alice!",
		},
		{
			name:     "unknown tag kept",
			template: []byte("Hello {{unknown}}!"),
			startTag: []byte("{{"),
			endTag:   []byte("}}"),
			values:   map[string]any{"name": "Alice"},
			want:     "Hello {{unknown}}!",
		},
		{
			name:     "nil value tag omitted",
			template: []byte("Hello {{empty}}!"),
			startTag: []byte("{{"),
			endTag:   []byte("}}"),
			values:   map[string]any{"empty": nil},
			want:     "Hello !",
		},
		{
			name:     "byte slice value",
			template: []byte("Data: {{data}}"),
			startTag: []byte("{{"),
			endTag:   []byte("}}"),
			values:   map[string]any{"data": []byte("binary")},
			want:     "Data: binary",
		},
		{
			name:     "nested tag function",
			template: []byte("Func: {{dynamic}}"),
			startTag: []byte("{{"),
			endTag:   []byte("}}"),
			values: map[string]any{
				"dynamic": TagFunc(func(w io.Writer, tag []byte) (int, error) {
					return w.Write([]byte("computed"))
				}),
			},
			want: "Func: computed",
		},
		{
			name:     "tag at start and end",
			template: []byte("{{start}} middle {{end}}"),
			startTag: []byte("{{"),
			endTag:   []byte("}}"),
			values:   map[string]any{"start": ">>", "end": "<<"},
			want:     ">> middle <<",
		},
		{
			name:     "no tags",
			template: []byte("Just a normal string"),
			startTag: []byte("{{"),
			endTag:   []byte("}}"),
			values:   map[string]any{},
			want:     "Just a normal string",
		},
		{
			name:     "empty template",
			template: []byte(""),
			startTag: []byte("{{"),
			endTag:   []byte("}}"),
			values:   map[string]any{},
			want:     "",
		},
		{
			name:     "only start tag",
			template: []byte("Hello {{name"),
			startTag: []byte("{{"),
			endTag:   []byte("}}"),
			values:   map[string]any{"name": "Alice"},
			want:     "Hello {{name",
		},
		{
			name:     "only end tag",
			template: []byte("Hello name}}"),
			startTag: []byte("{{"),
			endTag:   []byte("}}"),
			values:   map[string]any{"name": "Alice"},
			want:     "Hello name}}",
		},
		{
			name:     "different delimiters",
			template: []byte("Hello <%name%>!"),
			startTag: []byte("<%"),
			endTag:   []byte("%>"),
			values:   map[string]any{"name": "Bob"},
			want:     "Hello Bob!",
		},
		{
			name:     "strval on unexpected type",
			template: []byte("Hello {{age}}!"),
			startTag: []byte("{{"),
			endTag:   []byte("}}"),
			values:   map[string]any{"age": 30},
			want:     "Hello 30!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			_, err := EvalTagFunc(&buf, EvalTagFuncOptions{
				Template: tt.template,
				StartTag: tt.startTag,
				EndTag:   tt.endTag,
				Func:     KeepUnknownTags(tt.startTag, tt.endTag, tt.values),
			})

			assert.Nil(t, err)
			assert.Equal(t, tt.want, buf.String())
		})
	}
}
