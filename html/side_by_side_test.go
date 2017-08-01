package html

import (
	"fmt"
	"github.com/yu-ichiko/go-diff2html/diff"
	"testing"
)

func TestNewSideBySide(t *testing.T) {
	input := "--- a/sample.js\n" +
		"+++ b/sample.js\n" +
		"@@ -1 +1,2 @@\n" +
		"-test\n" +
		"+test1r\n" +
		"+test2r\n"
	d := diff.New(diff.Config{
		SrcPrefix: "",
		DstPrefix: "",
	})
	d.Parser(input)

	conf := Config{}
	side := NewSideBySide(conf)
	html, err := side.GenerateSideBySideHTML(d.Files)
	fmt.Println(html)
	fmt.Println(err)
}

func TestSideBySidePrinter_makeSideHTML(t *testing.T) {
	conf := Config{}
	side := NewSideBySide(conf)
	html, err := side.makeSideHTML("header")
	fmt.Println(html)
	fmt.Println(err)
}

func TestSideBySidePrinter_genSingleLineHTML(t *testing.T) {
	conf := Config{}
	side := NewSideBySide(conf)
	html, err := side.genSingleLineHTML(false, "d2h-cntx", 1, "{", " ")
	fmt.Println(html)
	fmt.Println(err)
}

func TestSideBySidePrinter_genEmptyDiff(t *testing.T) {
	conf := Config{}
	side := NewSideBySide(conf)
	fileHTML, err := side.genEmptyDiff()
	fmt.Println(fileHTML.Left)
	fmt.Println(err)
}

func TestSideBySidePrinter_processLines(t *testing.T) {
	oldLine := make([]*diff.Line, 3)
	newLine := make([]*diff.Line, 5)

	conf := Config{}
	side := NewSideBySide(conf)
	side.processLines(true, oldLine, newLine)
}

func Test_getDiffName(t *testing.T) {
	file := &diff.File{
		OldName: "sample",
		NewName: "sample2",
	}
	name := getDiffName(file)
	fmt.Println(name)
}

func Test_diffHighlight(t *testing.T) {
	highlight := diffHighlight(" category:campaign,", " category:guidance,", false)
	fmt.Println(highlight)
}

func Test_hoge(t *testing.T) {
	arr := []int{1, 2, 3, 4, 5}
	fmt.Println(arr)
	fmt.Println(arr[2:])
	fmt.Println(arr)
}
