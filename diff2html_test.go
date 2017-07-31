package diff2html

import (
	"testing"
	"fmt"
)

func Test_GetPrettyHTML(t *testing.T) {
	diff := "diff --git a/sample b/sample\n" +
		"index 0000001..0ddf2ba\n" +
		"--- a/sample\n" +
		"+++ b/sample\n" +
		"@@ -1 +1 @@\n" +
		"-test\n" +
		"+test1r\n"
	html, err := GetPrettyHTML(diff)
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
