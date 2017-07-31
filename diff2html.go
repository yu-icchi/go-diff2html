package diff2html

import (
	"github.com/yu-ichiko/go-diff2html/diff"
	"github.com/yu-ichiko/go-diff2html/html"
)

// GetPrettyHTML Generates the html diff.
func GetPrettyHTML(input string) (string, error) {
	d := diff.New(diff.Config{
		DstPrefix: "",
		SrcPrefix: "",
	})
	err := d.Parser(input)
	if err != nil {
		return "", err
	}
	diffHTML := html.NewSideBySide(html.Config{})
	return diffHTML.GenerateSideBySideHTML(d.Files)
}
