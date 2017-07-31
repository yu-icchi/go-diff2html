package diff2html

import (
	"github.com/yu-ichiko/go-diff2html/diff"
)

// GetPrettyHTML Generates the html diff.
func GetPrettyHTML(input string) (string, error) {
	conf := diff.Config{
		DstPrefix: "",
		SrcPrefix: "",
	}
	d := diff.New(conf)
	d.Parser(input)
	return "", nil
}
