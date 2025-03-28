package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarks(t *testing.T) {
	tests := []struct {
		name      string
		template  []byte
		startTag  []byte
		endTag    []byte
		want      []string
		wantError bool
	}{
		{
			name:     "single tag",
			template: []byte("Hello {{name}}!"),
			startTag: []byte("{{"),
			endTag:   []byte("}}"),
			want:     []string{"name"},
		},
		{
			name:     "multiple tags",
			template: []byte("{{greet}}, {{name}}!"),
			startTag: []byte("{{"),
			endTag:   []byte("}}"),
			want:     []string{"greet", "name"},
		},
		{
			name:     "nested-like but separate tags",
			template: []byte("{{outer}} {{inner}}"),
			startTag: []byte("{{"),
			endTag:   []byte("}}"),
			want:     []string{"outer", "inner"},
		},
		{
			name:     "tag with spaces",
			template: []byte("Hello {{ full name }}!"),
			startTag: []byte("{{"),
			endTag:   []byte("}}"),
			want:     []string{"full name"},
		},
		{
			name:     "tag at start and end",
			template: []byte("{{start}} middle {{end}}"),
			startTag: []byte("{{"),
			endTag:   []byte("}}"),
			want:     []string{"start", "end"},
		},
		{
			name:     "no tags",
			template: []byte("Just a normal string"),
			startTag: []byte("{{"),
			endTag:   []byte("}}"),
			want:     []string{},
		},
		{
			name:     "empty template",
			template: []byte(""),
			startTag: []byte("{{"),
			endTag:   []byte("}}"),
			want:     []string{},
		},
		{
			name:     "only start tag",
			template: []byte("Hello {{name"),
			startTag: []byte("{{"),
			endTag:   []byte("}}"),
			want:     []string{},
		},
		{
			name:     "only end tag",
			template: []byte("Hello name}}"),
			startTag: []byte("{{"),
			endTag:   []byte("}}"),
			want:     []string{},
		},
		{
			name:     "different delimiters",
			template: []byte("Hello <%name%>!"),
			startTag: []byte("<%"),
			endTag:   []byte("%>"),
			want:     []string{"name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marks(tt.template, tt.startTag, tt.endTag)
			if err != nil {
				assert.Equal(t, tt.wantError, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
