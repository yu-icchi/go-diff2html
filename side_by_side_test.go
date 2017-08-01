package diff2html

import (
	"fmt"
	"testing"
)

func TestNewSideBySide(t *testing.T) {
	input := "--- a/sample.js\n" +
		"+++ b/sample.js\n" +
		"@@ -1 +1,2 @@\n" +
		"-test\n" +
		"+test1r\n" +
		"+test2r\n"
	d := newDiff(Config{
		SrcPrefix: "",
		DstPrefix: "",
	})
	d.Parser(input)

	side := newSideBySide()
	html, err := side.GenerateSideBySideHTML(d.Files)
	fmt.Println(html)
	fmt.Println(err)
}

func TestSideBySidePrinter_makeSideHTML(t *testing.T) {
	side := newSideBySide()
	html, err := side.makeSideHTML("header")
	fmt.Println(html)
	fmt.Println(err)
}

func TestSideBySidePrinter_genSingleLineHTML(t *testing.T) {
	side := newSideBySide()
	html, err := side.genSingleLineHTML(false, "d2h-cntx", 1, "{", " ")
	fmt.Println(html)
	fmt.Println(err)
}

func TestSideBySidePrinter_genEmptyDiff(t *testing.T) {
	side := newSideBySide()
	fileHTML, err := side.genEmptyDiff()
	fmt.Println(fileHTML.Left)
	fmt.Println(err)
}

func TestSideBySidePrinter_processLines(t *testing.T) {
	oldLine := make([]*Line, 3)
	newLine := make([]*Line, 5)

	side := newSideBySide()
	side.processLines(true, oldLine, newLine)
}

func Test_getDiffName(t *testing.T) {
	file := &File{
		OldName: "sample",
		NewName: "sample2",
	}
	name := getDiffName(file)
	fmt.Println(name)
}

func Test_diffHighlight(t *testing.T) {
	highlight := diffHighlight(" category:campaign,", " category:guidance,", false)
	fmt.Println(highlight.First.Line)
	fmt.Println(highlight.Second.Line)
}
