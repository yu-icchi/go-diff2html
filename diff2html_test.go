package diff2html

import (
	"testing"
)

func Test_GetPrettyHTML(t *testing.T) {
	diff := "diff --git a/sample b/sample\n" +
		"index 0000001..0ddf2ba\n" +
		"--- a/sample\n" +
		"+++ b/sample\n" +
		"@@ -1 +1 @@\n" +
		"-test\n" +
		"+test1r\n";
	GetPrettyHTML(diff)
}
