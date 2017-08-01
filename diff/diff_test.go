package diff

import (
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	diff := "diff --git a/sample b/sample\n" +
		"index 0000001..0ddf2ba\n" +
		"--- a/sample\n" +
		"+++ b/sample\n" +
		"@@ -1 +1 @@\n" +
		"-test\n" +
		"+test1r\n"
	conf := Config{
		SrcPrefix: "",
		DstPrefix: "",
	}
	d := New(conf)
	d.Parser(diff)

	fmt.Println(d.Files)
}

func TestNew2(t *testing.T) {
	diff := "diff --git a/sample b/sample\r\n" +
		"index 0000001..0ddf2ba\r\n" +
		"--- a/sample\r\n" +
		"+++ b/sample\r\n" +
		"@@ -1 +1 @@\r\n" +
		"-test\r\n" +
		"+test1r\r\n"
	conf := Config{
		SrcPrefix: "",
		DstPrefix: "",
	}
	d := New(conf)
	d.Parser(diff)

	fmt.Println(d.Files)
}

func TestNew3(t *testing.T) {
	diff := "--- a/sample.js\n" +
		"+++ b/sample.js\n" +
		"@@ -1 +1,2 @@\n" +
		"-test\n" +
		"+test1r\n" +
		"+test2r\n"
	conf := Config{
		SrcPrefix: "",
		DstPrefix: "",
	}
	d := New(conf)
	d.Parser(diff)

	fmt.Println(d.Files)
}
