package template

import "io"

// Marks returns the list of all placeholders found in the specified template.
func Marks(template, startTag, endTag []byte) ([]string, error) {
	list := []string{}

	_, err := EvalTagFunc(io.Discard, EvalTagFuncOptions{
		Template: template,
		StartTag: startTag,
		EndTag:   endTag,
		Func:     CollectTags(&list),
	})

	return list, err
}
