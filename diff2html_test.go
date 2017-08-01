package diff2html

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pmezard/go-difflib/difflib"
	"testing"
)

func Test_GetPrettyHTML(t *testing.T) {

	a, _ := json.Marshal(struct {
		A string `json:"a"`
		B int    `json:"b"`
	}{
		A: "cat",
		B: 100,
	})
	var ao bytes.Buffer
	json.Indent(&ao, a, "", "\n")

	b, _ := json.Marshal(struct {
		A string `json:"a"`
		B int    `json:"b"`
	}{
		A: "doc",
		B: 10,
	})
	var bo bytes.Buffer
	json.Indent(&bo, b, "", "\n")

	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(ao.String()),
		B:        difflib.SplitLines(bo.String()),
		FromFile: "Original",
		ToFile:   "Current",
		Context:  3,
	}
	input, _ := difflib.GetUnifiedDiffString(diff)
	fmt.Printf(input)

	html, err := GetPrettyHTML(input)
	fmt.Println(html)
	fmt.Println(err)
}

func BenchmarkGetPrettyHTML(b *testing.B) {
	diff := "diff --git a/sample b/sample\n" +
		"index 0000001..0ddf2ba\n" +
		"--- a/sample\n" +
		"+++ b/sample\n" +
		"@@ -1 +1 @@\n" +
		"-test\n" +
		"+test1r\n"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetPrettyHTML(diff)
	}
}

// 5000	    226417 ns/op	  280411 B/op	     881 allocs/op
